// Adapted from https://github.com/TTK4145/driver-go

package door

import "elevatorinterface"

func Toggle(value bool) {
	elevatorinterface.Write([4]byte{4, elevatorinterface.ToByte(value), 0, 0})
}
