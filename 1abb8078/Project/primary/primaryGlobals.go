package primary

import (
	"Project/elevalgo"
	"Project/elevio"
	"time"
)

// package scope variables
var elevators = make(map[int]elevalgo.Elevator)// indexed by id

var hallLightMatrix [elevio.NUM_FLOORS][elevio.NUM_BUTTONS - 1]bool

var hallRequests [elevio.NUM_FLOORS][elevio.NUM_BUTTONS - 1]hallRequestsAndTime
var cabRequests = make(map[int][elevio.NUM_FLOORS]bool)

var lastTimeSpawnBackupSent time.Time

var onlyThisMachine bool 
var hasBeenConnectedOnce bool

// data types 
type hallRequestsAndTime struct {
	active bool 
	id int
	timeAdded time.Time
}
type timedOutHallRequest struct {
	id int
	floor int
	btn int
}
type requestWithId struct {
	id int
	btnEvent elevio.ButtonEvent
}
type elevatorConnectedWithId struct {
	id int
}
type elevatorAliveWithId struct {
	id int
	elevator elevalgo.Elevator
}
