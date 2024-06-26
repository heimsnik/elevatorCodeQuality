package singleelev

import (
	"encoding/json"
	"time"

	"Driver-go/elevio"

	master_slave "elevator/master-slave"
	"elevator/structs"
	tcp_interface "elevator/tcp-interface"
)

type Elevator struct {
	// The buffer values received from the elevio interface
	button_order  *elevio.ButtonEvent
	floor_sensor  *int
	is_obstructed *bool
	is_stopped    *bool

	// Variable containing the current state
	internal_state *structs.ElevatorState

	// Variable showing the last visited floor
	at_floor *int

	// The current target of the elevator (-1 for no target)
	target_floor *int

	// Variable for the direction of the elevator
	moving_direction *structs.Direction

	// Variable for keeping track of when interrupt ends
	interrupt_end *time.Time
	ms_unit       *master_slave.MasterSlave
}

func MakeElevator(elevatorNumber int, master *master_slave.MasterSlave) Elevator {
	// Starting values
	var start_state structs.ElevatorState = structs.IDLE
	starting_floor := -1
	target_floor := -1
	starting_direction := structs.STILL

	// Initial variables reading from channels
	floor_number := -1
	is_obstructed := false
	is_stopped := false

	start_time := time.Now()

	return Elevator{
		&elevio.ButtonEvent{},
		&floor_number,
		&is_obstructed,
		&is_stopped,
		&start_state,
		&starting_floor,
		&target_floor,
		&starting_direction,
		&start_time,
		master}
}

func (e Elevator) ElevatorLoop() {

	// Run at start to find correct floor
	if *e.at_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		*e.internal_state = structs.MOVING
	}

	for {

		// Always check for stop-button press
		if *e.is_stopped {
			*e.internal_state = structs.STOPPED
			e.AddElevatorDataToMaster()
		}

		switch state := *e.internal_state; state {
		case structs.IDLE:

			// Either move to existing target or choose new target
			if *e.at_floor != -1 {
				// Pick new target if no target, and the floor of the elevator is known
				e.PickTarget()
				time.Sleep(10 * time.Millisecond)
			}

		case structs.MOVING:
			// Run when arriving at new floor or when starting from target floor
			if *e.floor_sensor != -1 {
				e.PickTarget()
				e.MoveToTarget()
			}

			if (*e.at_floor != *e.floor_sensor || *e.floor_sensor == *e.target_floor) && *e.floor_sensor != -1 {

				// Set correct floor if not in between floors
				*e.at_floor = *e.floor_sensor

				// Update value of master
				e.AddElevatorDataToMaster()

				elevio.SetFloorIndicator(*e.at_floor)

				// Run visit floor routine
				e.Visit_floor()
				continue
			}

		case structs.DOOR_OPEN:
			e.OpenDoor()

		case structs.STOPPED:
			e.Stop()

		case structs.OBSTRUCTED:
			e.Obstruct()
		}
		
		// Keep the main loop from running unnecesarily fast
		time.Sleep(10 * time.Millisecond)
	}
}

// Read from the channels and put data into corresponding variables
func (e Elevator) ReadChannels(button_order chan elevio.ButtonEvent, current_floor chan int, is_obstructed chan bool, is_stopped chan bool) {

	for {
		select {
		case bo := <-button_order:
			// Read order and update system
			floor, btn := e.InterpretOrder(bo)
			e.AddOrderToSystemDAta(floor, btn)

		case cf := <-current_floor:
			*e.floor_sensor = cf

		case io := <-is_obstructed:
			*e.is_obstructed = io

		case is := <-is_stopped:
			*e.is_stopped = is
		}
	}
}

// Unpack button event
func (e *Elevator) InterpretOrder(button_order elevio.ButtonEvent) (floor int, button elevio.ButtonType) {
	order_floor := button_order.Floor
	order_button := button_order.Button

	return order_floor, order_button
}

// Add order to system data
func (e *Elevator) AddOrderToSystemDAta(floor int, button elevio.ButtonType) {

	switch button {
	case 0:
		e.AddHallOrderToMaster(floor, button)
	case 1:
		e.AddHallOrderToMaster(floor, button)
	case 2:
		e.AddCabOrderToMaster(floor)
	}
}

// Find the most suitable target for the elevator
func (e *Elevator) PickTarget() {
	self := e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[e.ms_unit.UNIT_ID]

	up_calls := [structs.N_FLOORS]bool{false, false, false, false}
	down_calls := [structs.N_FLOORS]bool{false, false, false, false}

	// Get calls in up and down directions
	targets := self.ELEVATOR_TARGETS
	for i := 0; i < structs.N_FLOORS; i++ {
		up_calls[i] = targets[i][0]
		down_calls[i] = targets[i][1]
	}

	// Get cab calls
	cab_calls := self.INTERNAL_BUTTON_ARRAY

	// New target floor and value to signify that it has been changed
	new_target := *e.target_floor
	updated := false

	for i := 0; i < structs.N_FLOORS; i++ {
		if *e.at_floor+i < structs.N_FLOORS {

			check_floor := *e.at_floor + i

			// Return if order is out of bound, or if elevator is moving in oposite direction
			if check_floor < 0 || check_floor > 4 || *e.moving_direction == structs.DOWN {
				continue
			}

			// Set target if an order exists on floor
			// Check if elevator is still or going upwards

			// Only move to down-calls when staying still
			down_when_still := down_calls[check_floor] && (*e.moving_direction == structs.STILL)
			// Always serve up calls and cab_calls
			if up_calls[check_floor] || cab_calls[check_floor] || down_when_still {
				new_target = check_floor
				updated = true
				break
			}

		}
		if *e.at_floor-i >= 0 {
			// Check floors below
			check_floor := *e.at_floor - i

			// Return if order is out of bound, or if elevator is moving in oposite direction
			if check_floor < 0 || check_floor > 4 || *e.moving_direction == structs.UP {
				continue
			}

			// Only move to down-calls when staying still
			up_when_still := up_calls[check_floor] && (*e.moving_direction == structs.STILL)
			if down_calls[check_floor] || cab_calls[check_floor] || up_when_still {
				new_target = check_floor
				updated = true
				break
			}
		}
	}

	// Run if a target has been found, and it is not the same as the previous target
	if updated && new_target != *e.target_floor {
		*e.target_floor = new_target

		// Begin moving towards target
		e.MoveToTarget()
		e.AddElevatorDataToMaster()
	}

}

// Check if door should be opened when visiting floor
func (e Elevator) Visit_floor() {

	// The only time the code reaches this state is during initialization
	if *e.target_floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		*e.internal_state = structs.IDLE
		e.AddElevatorDataToMaster()
		return
	}

	if *e.at_floor == *e.target_floor {

		// Reset target
		*e.target_floor = -1
		*e.moving_direction = structs.STILL

		// Transition to OpenDoor state
		elevio.SetMotorDirection(elevio.MD_Stop)
		*e.internal_state = structs.DOOR_OPEN

		// Add data to master
		e.ClearOrderFromMaster(*e.at_floor, *e.moving_direction)
		e.AddElevatorDataToMaster()

	}

}

func (e Elevator) OpenDoor() {

	elevio.SetDoorOpenLamp(true)

	// Open door
	time.Sleep(3 * time.Second)

	// Enter obstruction state if obstructed
	obstruction_check := *e.is_obstructed
	if obstruction_check {
		*e.internal_state = structs.OBSTRUCTED
		e.AddElevatorDataToMaster()
		return
	}

	// Close door
	elevio.SetDoorOpenLamp(false)

	// Return to idle
	*e.internal_state = structs.IDLE
	e.AddElevatorDataToMaster()

}

// Handles the obstructed state
func (e Elevator) Obstruct() {
	elevator_obstruct := *e.is_obstructed

	if !elevator_obstruct {
		// Begin timer when no longer obstructed
		time.Sleep(3 * time.Second)
		*e.internal_state = structs.DOOR_OPEN
		e.AddElevatorDataToMaster()
	}
}

// Sends the elevator in the right direction towards target
func (e Elevator) MoveToTarget() {
	// Set state to MOVING and set motor direction
	*e.internal_state = structs.MOVING

	if *e.target_floor > *e.at_floor {
		*e.moving_direction = structs.UP
		elevio.SetMotorDirection(elevio.MD_Up)
	} else if *e.target_floor < *e.at_floor {
		*e.moving_direction = structs.DOWN
		elevio.SetMotorDirection(elevio.MD_Down)
	}

}

// Handles stopping state
func (e Elevator) Stop() {

	// Check if stop button is pressed
	elevator_stop := *e.is_stopped

	// Stop elevator and activate lights
	elevio.SetStopLamp(true)
	elevio.SetMotorDirection(elevio.MD_Stop)

	if !elevator_stop {
		// Start timer when stop button is released
		time.Sleep(3 * time.Second)

		// Return to idle
		*e.internal_state = structs.MOVING
		e.AddElevatorDataToMaster()

		// Deactivate lights
		elevio.SetStopLamp(false)
		elevio.SetDoorOpenLamp(false)
	}
}

// Add cab order to data in master unit
func (e Elevator) AddCabOrderToMaster(floor int) {
	// Encode data
	data := structs.HallorderMsg{
		Order_floor:     floor,
		Order_direction: [2]bool{false, false},
	}
	encoded_data, _ := json.Marshal(&data)

	e._message_data_to_master(encoded_data, structs.NEWCABCALL)
}

// Add hall order to data in master unit
func (e Elevator) AddHallOrderToMaster(floor int, button elevio.ButtonType) {
	// Translate order to correct format
	dir_bool := [2]bool{false, false}
	if button == elevio.BT_HallUp {
		dir_bool[0] = true
	}
	if button == elevio.BT_HallDown {
		dir_bool[1] = true
	}

	// Encode data
	data := structs.HallorderMsg{
		Order_floor:     floor,
		Order_direction: dir_bool,
	}
	encoded_data, _ := json.Marshal(&data)

	e._message_data_to_master(encoded_data, structs.NEWHALLORDER)
}

// Update the elevator data stored in the master unit
func (e *Elevator) AddElevatorDataToMaster() {
	// Encode data
	data_copy := e.ms_unit.CURRENT_DATA
	id := e.ms_unit.UNIT_ID
	data_copy.ELEVATOR_DATA[id].CURRENT_FLOOR = *e.at_floor
	data_copy.ELEVATOR_DATA[id].DIRECTION = *e.moving_direction
	data_copy.ELEVATOR_DATA[id].INTERNAL_STATE = *e.internal_state

	encoded_data := tcp_interface.EncodeSystemData(data_copy)

	// Send data to master
	e._message_data_to_master(encoded_data, structs.UPDATEELEVATOR)
}

// Sends message to master to remove a specified order
func (e Elevator) ClearOrderFromMaster(floor int, dir structs.Direction) {

	clear_direction := [2]bool{false, false}

	if dir == structs.UP {
		clear_direction[0] = true
	} else if dir == structs.DOWN {
		clear_direction[1] = true
	} else if dir == structs.STILL {
		clear_direction[0] = true
		clear_direction[1] = true
	}
	// Encode data
	data := structs.HallorderMsg{
		Order_floor:     floor,
		Order_direction: clear_direction,
	}

	encoded_data, _ := json.Marshal(&data)

	e._message_data_to_master(encoded_data, structs.CLEARHALLORDER)
}

// Send a tcp-message with data to the master unit
func (e Elevator) _message_data_to_master(data []byte, msg_type structs.MessageType) {
	master_id := e.ms_unit.CURRENT_DATA.MASTER_ID

	// Construct message
	msg := structs.TCPMsg{
		MessageType: msg_type,
		Sender_id:   e.ms_unit.UNIT_ID,
		Data:        data,
	}
	encoded_msg := tcp_interface.EncodeMessage(&msg)

	// Send message to master
	tcp_interface.SendData(e.ms_unit.CURRENT_DATA.ELEVATOR_DATA[master_id].ADDRESS, encoded_msg)
}

// Reset all elevator elements
func ResetElevator() {
	// Set motor direction to stop
	elevio.SetMotorDirection(elevio.MD_Stop)

	// Turn off stop lamb
	elevio.SetStopLamp(false)

	// Turn of open door lamp
	elevio.SetDoorOpenLamp(false)

	// Reset all order lights
	for f := 0; f < structs.N_FLOORS; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
}
