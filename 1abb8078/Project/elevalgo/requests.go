package elevalgo

import (
	"Project/elevio"
	"fmt"
)

type DirnBehaviourPair struct {
	Dirn elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

func req_above(e Elevator) bool {
	for f := e.Floor + 1; f < elevio.NUM_FLOORS; f++ {
		for btn := 0; btn < elevio.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func req_below(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func req_here(e Elevator) bool {
	for btn := 0; btn < elevio.NUM_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}


func Req_chooseDirection(e Elevator) DirnBehaviourPair {
    switch e.Dirn {
    case elevio.MD_Up:
        if req_above(e) {
            return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
        } else if req_here(e) {
            return DirnBehaviourPair{elevio.MD_Down, EB_DoorOpen}
        } else if req_below(e) {
            return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
        }
    case elevio.MD_Down:
        if req_below(e) {
            return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
        } else if req_here(e) {
            return DirnBehaviourPair{elevio.MD_Up, EB_DoorOpen}
        } else if req_above(e) {
            return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
        }
    case elevio.MD_Stop:
        if req_here(e) {
            return DirnBehaviourPair{elevio.MD_Stop, EB_DoorOpen}
        } else if req_above(e) {
            return DirnBehaviourPair{elevio.MD_Up, EB_Moving}
        } else if req_below(e) {
            return DirnBehaviourPair{elevio.MD_Down, EB_Moving}
        } else {
            return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
        }
    default:
        return DirnBehaviourPair{elevio.MD_Stop, EB_Idle}
    }
}


func Req_shouldStop(e Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return 	e.Requests[e.Floor][elevio.BT_HallDown] ||
				e.Requests[e.Floor][elevio.BT_Cab] 		|| 
				!req_below(e)
	case elevio.MD_Up:
		return 	e.Requests[e.Floor][elevio.BT_HallUp] 	|| 
				e.Requests[e.Floor][elevio.BT_Cab] 		|| 
				!req_above(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}
         
func Req_shouldClearImmediately(e Elevator, btn_Floor int, btn_type elevio.ButtonType) bool {
	return 	e.Floor == btn_Floor && 
			((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) 		|| 
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) 	|| 
			e.Dirn == elevio.MD_Stop 										|| 
			btn_type == elevio.BT_Cab)
}

func Req_clearAtCurrentFloor(e Elevator) Elevator {
	e.Requests[e.Floor][elevio.BT_Cab] = false
	switch e.Dirn {
	case elevio.MD_Up:
		if !req_above(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
		e.Requests[e.Floor][elevio.BT_HallUp] = false
	case elevio.MD_Down:
		if !req_below(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		}
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	case elevio.MD_Stop:
		e.Requests[e.Floor][elevio.BT_HallUp] = false
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	default:
		fmt.Println(" Req_clearAtCurrentFloor: Something very wrong")
	}
	return e
}

func Req_clearCabRequest(e Elevator, floor int) Elevator {
		e.Requests[floor][elevio.BT_Cab] = false
	return e
}

