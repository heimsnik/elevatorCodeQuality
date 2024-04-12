// Adapted from https://github.com/TTK4145/Project-resources/blob/master/elev_algo/requests.c

package decision

import (
	"elevatorglobals"
)

func hasOrdersAbove(cabState elevatorglobals.CabState, assignedOrders elevatorglobals.AssignedOrders) bool {
	for floor := cabState.Floor + 1; floor < elevatorglobals.FloorCount; floor++ {
		for button := 0; button < elevatorglobals.ButtonCount; button++ {
			if assignedOrders[floor][button] {
				return true
			}
		}
	}
	return false
}

func hasOrdersBelow(cabState elevatorglobals.CabState, assignedOrders elevatorglobals.AssignedOrders) bool {
	for floor := 0; floor < cabState.Floor; floor++ {
		for button := 0; button < elevatorglobals.ButtonCount; button++ {
			if assignedOrders[floor][button] {
				return true
			}
		}
	}
	return false
}

func hasOrdersHere(cabState elevatorglobals.CabState, assignedOrders elevatorglobals.AssignedOrders) bool {
	for button := 0; button < elevatorglobals.ButtonCount; button++ {
		if assignedOrders[cabState.Floor][button] {
			return true
		}
	}
	return false
}

func ChooseDirection(cabState elevatorglobals.CabState, assignedOrders elevatorglobals.AssignedOrders) (elevatorglobals.Direction, elevatorglobals.ElevatorBehaviour) {
	switch cabState.Direction {
	case elevatorglobals.Direction_Up:
		if hasOrdersAbove(cabState, assignedOrders) {
			return elevatorglobals.Direction_Up, elevatorglobals.ElevatorBehaviour_Moving
		} else if hasOrdersHere(cabState, assignedOrders) {
			return elevatorglobals.Direction_Down, elevatorglobals.ElevatorBehaviour_DoorOpen
		} else if hasOrdersBelow(cabState, assignedOrders) {
			return elevatorglobals.Direction_Down, elevatorglobals.ElevatorBehaviour_Moving
		} else {
			return elevatorglobals.Direction_Stop, elevatorglobals.ElevatorBehaviour_Idle
		}

	case elevatorglobals.Direction_Down:
		if hasOrdersBelow(cabState, assignedOrders) {
			return elevatorglobals.Direction_Down, elevatorglobals.ElevatorBehaviour_Moving
		} else if hasOrdersHere(cabState, assignedOrders) {
			return elevatorglobals.Direction_Up, elevatorglobals.ElevatorBehaviour_DoorOpen
		} else if hasOrdersAbove(cabState, assignedOrders) {
			return elevatorglobals.Direction_Up, elevatorglobals.ElevatorBehaviour_Moving
		} else {
			return elevatorglobals.Direction_Stop, elevatorglobals.ElevatorBehaviour_Idle
		}

	case elevatorglobals.Direction_Stop:
		if hasOrdersHere(cabState, assignedOrders) {
			return elevatorglobals.Direction_Stop, elevatorglobals.ElevatorBehaviour_DoorOpen
		} else if hasOrdersAbove(cabState, assignedOrders) {
			return elevatorglobals.Direction_Up, elevatorglobals.ElevatorBehaviour_Moving
		} else if hasOrdersBelow(cabState, assignedOrders) {
			return elevatorglobals.Direction_Down, elevatorglobals.ElevatorBehaviour_Moving
		} else {
			return elevatorglobals.Direction_Stop, elevatorglobals.ElevatorBehaviour_Idle
		}
	}

	return elevatorglobals.Direction_Stop, elevatorglobals.ElevatorBehaviour_Idle
}

func ShouldStop(cabState elevatorglobals.CabState, assignedOrders elevatorglobals.AssignedOrders) bool {
	switch cabState.Direction {
	case elevatorglobals.Direction_Down:
		return assignedOrders[cabState.Floor][elevatorglobals.ButtonType_HallDown] ||
			assignedOrders[cabState.Floor][elevatorglobals.ButtonType_Cab] ||
			!hasOrdersBelow(cabState, assignedOrders)
	case elevatorglobals.Direction_Up:
		return assignedOrders[cabState.Floor][elevatorglobals.ButtonType_HallUp] ||
			assignedOrders[cabState.Floor][elevatorglobals.ButtonType_Cab] ||
			!hasOrdersAbove(cabState, assignedOrders)
	case elevatorglobals.Direction_Stop:
		return true
	default:
		return true
	}
}

func GetOrdersHandled(cabState elevatorglobals.CabState, assignedOrders elevatorglobals.AssignedOrders) []elevatorglobals.OrderEvent {
	handledCab := true
	handledUp := false
	handledDown := false
	switch cabState.Direction {
	case elevatorglobals.Direction_Up:
		if !hasOrdersAbove(cabState, assignedOrders) && !assignedOrders[cabState.Floor][elevatorglobals.ButtonType_HallUp] {
			handledDown = true
		}
		handledUp = true
	case elevatorglobals.Direction_Down:
		if !hasOrdersBelow(cabState, assignedOrders) && !assignedOrders[cabState.Floor][elevatorglobals.ButtonType_HallDown] {
			handledUp = true
		}
		handledDown = true
	case elevatorglobals.Direction_Stop:
		// Currently handles both hall up and down at the same time if elevator is stopped.
		// This may be unintentional behaviour, but it's the way requests.c implements it, so we keep it that way
		handledUp = true
		handledDown = true
	}

	ordersHandledList := make([]elevatorglobals.OrderEvent, 0)
	if handledCab {
		ordersHandledList = append(ordersHandledList, elevatorglobals.OrderEvent{OriginName: elevatorglobals.MyElevatorName, Floor: cabState.Floor, Button: elevatorglobals.ButtonType_Cab})
	}
	if handledUp {
		ordersHandledList = append(ordersHandledList, elevatorglobals.OrderEvent{OriginName: elevatorglobals.MyElevatorName, Floor: cabState.Floor, Button: elevatorglobals.ButtonType_HallUp})
	}
	if handledDown {
		ordersHandledList = append(ordersHandledList, elevatorglobals.OrderEvent{OriginName: elevatorglobals.MyElevatorName, Floor: cabState.Floor, Button: elevatorglobals.ButtonType_HallDown})
	}

	return ordersHandledList
}
