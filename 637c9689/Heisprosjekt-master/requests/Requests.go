package requests

import (
	"Heisprosjekt/elevator"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/utils"
)

func Requests_ChooseDirection() elevator.DirectionBehaviourPair {

	switch utils.MyElev.Direction {
	case elevcons.MD_Up:
		if Requests_Above() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Up, elevcons.Moving)
		} else if Requests_Here() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Down, elevcons.Door_open)
		} else if Requests_Below() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Down, elevcons.Moving)
		} else {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Stop, elevcons.Idle)
		}
	case elevcons.MD_Down:
		if Requests_Below() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Down, elevcons.Moving)
		} else if Requests_Here() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Up, elevcons.Door_open)
		} else if Requests_Above() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Up, elevcons.Moving)
		} else {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Stop, elevcons.Idle)
		}
	case elevcons.MD_Stop:
		if Requests_Here() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Stop, elevcons.Door_open)
		} else if Requests_Above() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Up, elevcons.Moving)
		} else if Requests_Below() {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Down, elevcons.Moving)
		} else {
			return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Stop, elevcons.Idle)
		}
	default:
		return elevator.Elevator_GetDirectionBehaviourPair(elevcons.MD_Stop, elevcons.Idle)
	}
}

func Requests_Above() bool {

	for f := utils.MyElev.CurrentFloor + 1; f < elevcons.N_Floors; f++ {
		for btn := 0; btn < elevcons.N_Buttons; btn++ {
			if utils.MyElev.Requests[f][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func Requests_Below() bool {

	for f := 0; f < utils.MyElev.CurrentFloor; f++ {
		for btn := 0; btn < elevcons.N_Buttons; btn++ {
			if utils.MyElev.Requests[f][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func Requests_Here() bool {

	for btn := 0; btn < elevcons.N_Buttons; btn++ {
		if utils.MyElev.Requests[utils.MyElev.CurrentFloor][btn] == 1 {
			return true
		}
	}
	return false
}

func Request_CheckRequestList(floor int, btn elevcons.ButtonType) {

	if utils.MyElev.Requests[floor][btn] == 1 {
		if utils.MyElev.Status == elevcons.Primary {
			utils.Utils_AddToRequestMap([][3]int{{elevcons.TurnOffLight, floor, int(btn)}})
		} else if utils.MyElev.Status == elevcons.Backup {
			utils.Utils_AddToRequestMap([][3]int{{elevcons.CompletedReq, floor, int(btn)}})
		}
	}
}

func Requests_ClearAtCurrentFloor() {

	utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_Cab] = 0
	utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_Cab] = 0

	switch utils.MyElev.Direction {
	case elevcons.MD_Up:
		if !Requests_Above() && !utils.Utils_IntToBool(utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallUp]) {
			Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallDown)
			utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
			utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
		}
		Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallUp)
		utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0
		utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0

	case elevcons.MD_Down:
		if !Requests_Below() && !utils.Utils_IntToBool(utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallDown]) {
			Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallUp)
			utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0
			utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0

		}
		Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallDown)
		utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
		utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0

	case elevcons.MD_Stop:
		Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallDown)
		Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallUp)
		utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
		utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0
		utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
		utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0

	default:
		Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallDown)
		Request_CheckRequestList(utils.MyElev.CurrentFloor, elevcons.BT_HallUp)
		utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
		utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0
		utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallDown] = 0
		utils.MyElev.Lights[utils.MyElev.CurrentFloor][elevcons.BT_HallUp] = 0

	}

}

func Requests_ShouldStop() bool {

	check1 := utils.Utils_IntToBool(utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallDown])
	check2 := utils.Utils_IntToBool(utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_Cab])
	check3 := utils.Utils_IntToBool(utils.MyElev.Requests[utils.MyElev.CurrentFloor][elevcons.BT_HallUp])

	switch utils.MyElev.Direction {
	case elevcons.MD_Down:
		return (check1 || check2 || !Requests_Below())

	case elevcons.MD_Up:
		return (check2 || check3 || !Requests_Above())

	case elevcons.MD_Stop:
		return true

	default:
		return true
	}
}

func Requests_ShouldClearImmediatly(btn_floor int, btn_type elevcons.ButtonType) bool {
	
	check1 := utils.MyElev.CurrentFloor == btn_floor
	check2 := (utils.MyElev.Direction == elevcons.MD_Up && btn_type == elevcons.BT_HallUp)
	check3 := (utils.MyElev.Direction == elevcons.MD_Down && btn_type == elevcons.BT_HallDown)
	check4 := utils.MyElev.Direction == elevcons.MD_Stop
	check5 := btn_type == elevcons.BT_Cab

	return (check1 && (check2 || check3 || check4 || check5))
}
