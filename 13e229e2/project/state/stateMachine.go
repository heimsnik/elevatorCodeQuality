package state

import (
	"Driver-go/config"
	"Driver-go/datatypes"
	"Driver-go/elevio"
	"Driver-go/network"
	"Driver-go/utilities"
	"fmt"
	"net"
	"time"
)

//////////// State machine data types

type State int

const (
	State_undefined State = iota
	State_idle
	State_moving
	State_floorStop
	State_fullStop
)

type Cab struct {
	currentState  State
	currentFloor  int
	connection    net.Conn
	direction     int
	obstruction   bool
	timerDone     bool
	id            int
	internalQueue []elevio.ButtonEvent
	isConnected   bool
	doneSend	  bool
}

//////////// Internal queueu functions

func (sm *Cab) Add_task(btn elevio.ButtonEvent) {
	sm.internalQueue = append(sm.internalQueue, btn)
	elevio.SetButtonLamp(btn.Button, btn.Floor, true)
}

func (sm *Cab) Remove_task(taskFloor int) elevio.ButtonEvent {
	var removed elevio.ButtonEvent
	var temp []elevio.ButtonEvent
	for _, element := range sm.internalQueue {
		if element.Floor != taskFloor {
			temp = append(temp, element)
		} else {
			removed = element
		}
	}
	sm.internalQueue = temp
	return removed
}

//////////// State machine functions

func Run_single_elev(elevId int) {
	elevio.Init("localhost:15657", config.NumFloors)
	sm := Cab{currentState: State_undefined, timerDone: false, id: elevId, isConnected: false, doneSend: false}
	Reset_all_btn_lamp(config.NumFloors)
	elevio.SetDoorOpenLamp(false)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	srv_listner := make(chan datatypes.ServerMessage)
	srv_connectionLoss := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//Try to find masters TCP server and if server is found connect to it
	masterExist, masterAddr := network.Find_server_ip()
	if masterExist {
		conn := network.Connect_to_server(masterAddr, "elevator")
		sm.connection = conn
		sm.isConnected = true
		fmt.Print("Connected to server")

		go network.Listen_to_server(srv_listner, srv_connectionLoss, conn)

		defer conn.Close()
	}

	doorTimer := time.NewTimer(config.DoorOpenDuration)
	doorTimer.Stop()

	fmt.Printf("Elevator with id %d is ready for use. \n", sm.id)

	for {
		select {
		case floor := <-drv_floors:
			sm.currentFloor = floor
			elevio.SetFloorIndicator(floor)
			//Send floor update to server
			msg := datatypes.ServerMessage{Tag: "updateFloor", Floor: floor}
			network.Send_message(msg, sm.connection)

		case btn := <-drv_buttons:
			if sm.isConnected {
				//Send to server
				msg := datatypes.ServerMessage{Tag: "newRequest", Task: btn}
				network.Send_message(msg, sm.connection)
			} else {
				//Else run local
				sm.Add_task(btn)
			}

		case <-drv_stop:
			sm.currentState = State_fullStop

		case obs := <-drv_obstr:
			sm.obstruction = obs

		case <-doorTimer.C:
			sm.timerDone = true

		case <-srv_connectionLoss:
			sm.isConnected = false
			//Try to reconnect to master
			masterExist, masterAddr := network.Find_server_ip()
			if masterExist {
				conn := network.Connect_to_server(masterAddr, "elevator")
				sm.connection = conn
				sm.isConnected = true

				go network.Listen_to_server(srv_listner, srv_connectionLoss, conn)
			}

		case newTask := <-srv_listner:
			if newTask.Tag == "task" {
				sm.Add_task(elevio.ButtonEvent{Floor: newTask.Task.Floor, Button: newTask.Task.Button})
			}

		default:
			sm.StateMachine(drv_floors, config.NumFloors, doorTimer)
		}
	}
}

func (sm *Cab) StateMachine(floor_sensor chan int, numFloors int, doorTimer *time.Timer) {
	switch sm.currentState {
	case State_undefined:
		sm.Elev_starting_routine(floor_sensor)

	case State_idle:
		doorTimer.Reset(config.DoorOpenDuration)
		if len(sm.internalQueue) > 0 {
			sm.currentState = State_moving
		}

	case State_moving:
		sm.doneSend = false
		doorTimer.Reset(config.DoorOpenDuration)
		//Take first task
		intermedietStop := sm.internalQueue[0].Floor
		//Check if there is a possible task between first task and current position of elevator
		if sm.internalQueue[0].Floor > sm.currentFloor {
			//Going up
			intermedietStop = utilities.Find_smallest(sm.internalQueue, sm.currentFloor)
		} else if sm.internalQueue[0].Floor < sm.currentFloor {
			//Going down
			intermedietStop = utilities.Find_largest(sm.internalQueue, sm.currentFloor)
		}
		sm.Go_to_floor(intermedietStop, doorTimer)

	case State_floorStop:
		if sm.direction == 0 {
			removed := sm.Remove_task(sm.currentFloor)
			Reset_btn_lamps_at_floor(sm.currentFloor)
			//If connected to server send "done" message
			if sm.isConnected && !sm.doneSend{
				msg := datatypes.ServerMessage{Tag: "done", Task: removed}
				network.Send_message(msg, sm.connection)
				sm.doneSend = true
			}
			sm.Open_door()
		}

	case State_fullStop:
		doorTimer.Reset(config.DoorOpenDuration)
		elevio.SetMotorDirection(elevio.MD_Stop)
		//sm.internalQueue = []elevio.ButtonEvent{}
		//Reset_all_btn_lamp(numFloors)
		utilities.Delay_ms(1000)
		sm.currentState = State_undefined
	}
}

func (sm *Cab) Elev_starting_routine(floor_sensor chan int) {
	elevio.SetMotorDirection(elevio.MD_Down)
	sm.currentFloor = <-floor_sensor

	elevio.SetFloorIndicator(sm.currentFloor)
	elevio.SetMotorDirection(elevio.MD_Stop)
	sm.direction = 0
	sm.currentState = State_idle
}

func (sm *Cab) Go_to_floor(target_floor int, doorTimer *time.Timer) {
	if sm.currentFloor == target_floor && elevio.GetFloor() != -1 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		sm.direction = 0
		sm.currentFloor = elevio.GetFloor()
		doorTimer.Reset(config.DoorOpenDuration)
		sm.currentState = State_floorStop
	} else if target_floor > sm.currentFloor {
		//Going up
		elevio.SetMotorDirection(elevio.MD_Up)
		sm.direction = 1
	} else if target_floor < sm.currentFloor {
		//Going down
		elevio.SetMotorDirection(elevio.MD_Down)
		sm.direction = -1
	}
	utilities.Delay_ms(10)
}

func (sm *Cab) Open_door() {
	elevio.SetDoorOpenLamp(true)

	if !sm.obstruction && sm.timerDone {
		sm.timerDone = false
		elevio.SetDoorOpenLamp(false)
		if len(sm.internalQueue) > 0 {
			sm.currentState = State_moving
		} else {
			sm.currentState = State_idle
		}
	}
	utilities.Delay_ms(10)
}

func Reset_btn_lamps_at_floor(floor int) {
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
}

func Reset_all_btn_lamp(floor_num int) {
	for floor := 0; floor < floor_num; floor++ {
		Reset_btn_lamps_at_floor(floor)
	}
}
