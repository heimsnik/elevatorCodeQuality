package clientElevator

import (
	"Project/network/tcpnetwork"
	"Project/elevio"
	"Project/elevalgo"
	"time"
	"fmt"
)

func pollPrimaryConnection(primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient, primaryState chan bool) {
	oldState := false
	for {
		time.Sleep(_pollConnectionCheckRate)
		newState := primary_socket.IsActive()
		if newState != oldState{
			fmt.Println("\n================================Client State Update - Connection is: ", newState, "=======================================")
			primaryState <- newState
			oldState = newState
		}
	}
}

func pollSendAlive(send_alive chan bool) {
	for {
		time.Sleep(_pollAliveRate)
		send_alive <- true
	}
}

func pollLights() {
	for {
		time.Sleep(_pollLightRate)
		//Set cab lights
		for i := 0; i < elevio.NUM_FLOORS; i++ {
			elevio.SetButtonLamp(elevio.BT_Cab, i, elevalgo.GetElevator().Requests[i][elevio.BT_Cab])
		}
		//Set hall lights
		if(isConnected){
			for j := 0; j < elevio.NUM_FLOORS; j++ {
				elevio.SetButtonLamp(elevio.BT_HallUp, j, primaryHallLights[j][elevio.BT_HallUp])
				elevio.SetButtonLamp(elevio.BT_HallDown, j, primaryHallLights[j][elevio.BT_HallDown])
			}
		}
	}
}
