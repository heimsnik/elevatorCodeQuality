package backup

import (
	"Project/elevalgo"
	"Project/elevio"
)

var elevators = make(map[int]elevalgo.Elevator) 
var confirmedHallRequests [elevio.NUM_FLOORS][elevio.NUM_BUTTONS-1]bool 
var confirmedCabRequests = make(map[int][elevio.NUM_FLOORS]bool)
