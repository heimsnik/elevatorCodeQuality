package clientElevator

import (
	"Project/elevio"
)

var isConnected bool
var primaryHallLights [elevio.NUM_FLOORS][elevio.NUM_BUTTONS - 1]bool