// Adapted from https://github.com/TTK4145/driver-go

package lamps

import (
	"elevatorglobals"
	"elevatorinterface"
)

func Update(worldview elevatorglobals.Worldview) {
	myElevatorIndex := worldview.ElevatorIndex(elevatorglobals.MyElevatorName)
	if myElevatorIndex == -1 {
		return
	}
	for floorIndex := range elevatorglobals.FloorCount {
		setButtonLamp(elevatorglobals.ButtonType_HallUp, floorIndex, worldview.HallOrders[floorIndex][int(elevatorglobals.ButtonType_HallUp)])
		setButtonLamp(elevatorglobals.ButtonType_HallDown, floorIndex, worldview.HallOrders[floorIndex][int(elevatorglobals.ButtonType_HallDown)])
		setButtonLamp(elevatorglobals.ButtonType_Cab, floorIndex, worldview.CabOrders[myElevatorIndex][floorIndex])
	}
}

func setButtonLamp(button elevatorglobals.ButtonType, floor int, value bool) {
	elevatorinterface.Write([4]byte{2, byte(button), byte(floor), elevatorinterface.ToByte(value)})
}
