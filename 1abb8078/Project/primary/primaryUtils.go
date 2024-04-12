package primary

import (
	"Project/elevio"
	"strings"
	"strconv"
	"time"
	"Project/network/udpnetwork"
	"fmt"
)

func updateHallLightMatrix() {
	for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
		for btn := 0; btn < elevio.NUM_BUTTONS-1; btn++ {
			hallLightMatrix[floor][btn] = hallRequests[floor][btn].active
		}
	}
}

func getMachineID(ip string) int {
	// Get machine ID from IP. We use last IP octet as ID for each elevator
	octets := strings.Split(ip, ".")
	lastOctet := octets[len(octets)-1]
	ID, _ := strconv.Atoi(lastOctet)
	return ID
}
func killIfOtherPrimaryBroadcast(){
	otherPrimaryUDPListner := udpnetwork.NewElevatorUDPClient()
	stopChan := make(chan bool)
	
	otherPrimaryUDPListner.ListenForUDPBroadcastedIP(udpnetwork.PRIMARY_ONLY_BROADCAST_PORT, stopChan)
	noOtherPrimaryTimer := time.NewTimer(_pollPrimaryAliveRate) 
	select {
		case <- otherPrimaryUDPListner.In:
			fmt.Println("PrimaryMain: Other primary found")
			stopChan <- true
			otherPrimaryUDPListner.Stop()
			return
		case <- noOtherPrimaryTimer.C:
			fmt.Println("PrimaryMain: No other primary found")
			stopChan <- true
			otherPrimaryUDPListner.Stop()
			return
		}
}

func resetAllTimesAndIDs(hallRequests [elevio.NUM_FLOORS][elevio.NUM_BUTTONS-1]bool) [elevio.NUM_FLOORS][elevio.NUM_BUTTONS-1]hallRequestsAndTime {

	toReturn := [elevio.NUM_FLOORS][elevio.NUM_BUTTONS-1] hallRequestsAndTime{}

	for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
		for btn := 0; btn < elevio.NUM_BUTTONS-1; btn++ {
			toReturn[floor][btn].id = -1 
			if hallRequests[floor][btn] {
				toReturn[floor][btn].active = true
				toReturn[floor][btn].timeAdded = time.Now()
			}
		}
	}

	return toReturn
}