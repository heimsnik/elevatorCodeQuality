// Adapted from https://github.com/TTK4145/driver-go

package buttons

import (
	"elevatorglobals"
	"elevatorinterface"
	"time"
)

func Poll(receiver chan<- elevatorglobals.Order) {
	prev := make([][elevatorglobals.ButtonCount]bool, elevatorglobals.FloorCount)
	for {
		time.Sleep(elevatorinterface.PollRate)
		for f := 0; f < elevatorglobals.FloorCount; f++ {
			for b := elevatorglobals.ButtonType(0); b < elevatorglobals.ButtonCount; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- elevatorglobals.Order{Floor: f, Button: elevatorglobals.ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func getButton(button elevatorglobals.ButtonType, floor int) bool {
	a := elevatorinterface.Read([4]byte{6, byte(button), byte(floor), 0})
	return elevatorinterface.ToBool(a[1])
}
