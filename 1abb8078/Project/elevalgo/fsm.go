package elevalgo

import (
	"Project/elevio"
	"time"
)

func Fsm_onInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Behaviour = EB_Moving
	elevator.Dirn = elevio.MD_Down
	elevator.doorOpenDuration_s = 3 * time.Second
}

func removedRequests(e1 Elevator, e2 Elevator) []elevio.ButtonEvent {

	var removedRequests []elevio.ButtonEvent

	for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
		for btn := elevio.BT_HallUp; btn <= elevio.BT_Cab; btn++ {
			if e1.Requests[floor][btn] && !e2.Requests[floor][btn] {
				removedRequests = append(removedRequests, elevio.ButtonEvent{Floor:floor, Button:btn})
			}
		}
	}

	return removedRequests
}

func Fsm_onRequestButtonPress(btnFloor int, btnType elevio.ButtonType, timer *ElevatorTimer) []elevio.ButtonEvent {
	
	var oldElevator Elevator

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if Req_shouldClearImmediately(elevator, btnFloor, btnType) {
			timer.Start(elevator.doorOpenDuration_s)
		} else {
			elevator.Requests[btnFloor][btnType] = true
		}
	case EB_Moving:
		elevator.Requests[btnFloor][btnType] = true
	case EB_Idle:
		elevator.Requests[btnFloor][btnType] = true
		pair := Req_chooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.Start(elevator.doorOpenDuration_s)
			oldElevator = elevator
			elevator = Req_clearAtCurrentFloor(elevator)
		case EB_Moving:
			elevio.SetMotorDirection(elevator.Dirn)
		case EB_Idle:
		}
	}
	
	return removedRequests(oldElevator, elevator)
}

func Fsm_onFloorArrival(newFloor int, timer *ElevatorTimer) []elevio.ButtonEvent  {

	elevator.Floor = newFloor
	var oldElevator Elevator

	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case EB_Moving:
		if Req_shouldStop(elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			oldElevator = elevator
			elevator = Req_clearAtCurrentFloor(elevator)
			timer.Start(elevator.doorOpenDuration_s)
			elevator.Behaviour = EB_DoorOpen
		}
	default:
	}
	
	return removedRequests(oldElevator, elevator)
}

func Fsm_onDoorTimeout(timer *ElevatorTimer) []elevio.ButtonEvent {

	var oldElevator Elevator
	
	switch elevator.Behaviour {
	case EB_DoorOpen:
		if elevator.Obstruction {
			timer.Start(elevator.doorOpenDuration_s)
			break
		}

		pair := Req_chooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case EB_DoorOpen:
			timer.Start(elevator.doorOpenDuration_s)
			oldElevator = elevator
			elevator = Req_clearAtCurrentFloor(elevator)
		case EB_Moving, EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
		}
	default:
		break
	}

	return removedRequests(oldElevator, elevator)
}

func Fsm_onObstruction(a bool) {
	elevator.Obstruction = a
}

func Fsm_onReconnectClearCabRequest(floor int) {
	Req_clearCabRequest(elevator, floor)
}