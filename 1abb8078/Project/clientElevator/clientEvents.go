package clientElevator

import (
	"Project/elevalgo"
	"Project/elevio"
	"Project/backup"
	"Project/network/messages"
	"Project/network/tcpnetwork"
	"Project/network/udpnetwork"
	"fmt"
)

func decodeMessage(bytes []byte, 
	dr chan messages.M_DoRequest, 
	hl chan messages.M_HallLights, 
	sb chan messages.M_SpawnBackup, 
	kill chan messages.M_KILL){
		
	messageArray := messages.SplitMessages(bytes)
	for _, segment := range messageArray{

		message := messages.BytesToMessage([]byte(segment))

		switch message.(type) {
		case messages.M_DoRequest:
			dr <- message.(messages.M_DoRequest)
		case messages.M_HallLights:
			hl <- message.(messages.M_HallLights)
		case messages.M_SpawnBackup:
			sb <- message.(messages.M_SpawnBackup)
		case messages.M_KILL:
			kill <- message.(messages.M_KILL)
		default:
			fmt.Println("client.decodeMessage: Unknown message type: ", segment)
		}
	}		
}

func event_doRequest(a messages.M_DoRequest, doorTimeoutTimer *elevalgo.ElevatorTimer, primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient){
	fmt.Println("client.event_doRequest:  Floor: ", a.Data.Floor, " Button: ", a.Data.Button)
	finRequest := elevalgo.Fsm_onRequestButtonPress(a.Data.Floor, a.Data.Button, doorTimeoutTimer)
	for i := 0; i < len(finRequest); i++ {
		// Send completed request to primary
		completed_request := messages.M_CompletedRequest{Data:finRequest[i]}
		primary_socket.Out <- messages.MessageToBytes(completed_request)
		fmt.Println(" -> Sending comfirmed request to Primary nr:", i+1,)
	}
}

func event_doHallLights(hallLights messages.M_HallLights){
	isConnected = true
	primaryHallLights = hallLights.Data
}

func event_spawnBackup(udp_socket *udpnetwork.ElevatorUDPClient){
	fmt.Println("client.event_spawnBackup")
	backup.BackupMain(udp_socket.In)
}

func event_doKill(primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient, udp_socket *udpnetwork.ElevatorUDPClient){
	fmt.Println("client.event_doKill")
	primary_socket.Stop()
	udp_socket.Stop()
}

func event_sendAlive(primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient){
	if (isConnected) {
		alive := messages.M_ElevatorAlive{Data: elevalgo.GetElevator()}
		primary_socket.Out <- messages.MessageToBytes(alive)
	}
}

func event_buttonPress(btnEvent elevio.ButtonEvent, doorTimeoutTimer *elevalgo.ElevatorTimer, primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient){
	fmt.Println("client.event_buttonPress:  Floor: ", btnEvent.Floor, " Button: ", btnEvent.Button)
	if (isConnected){
		request := messages.M_NewRequest{Data: btnEvent}
		primary_socket.Out <- messages.MessageToBytes(request)
		fmt.Println(" -> Sending new request to Primary")
	} else {
		if(btnEvent.Button == elevio.BT_Cab){
			elevalgo.Fsm_onRequestButtonPress(btnEvent.Floor, btnEvent.Button, doorTimeoutTimer)
			fmt.Println(" -> Doing local cab request")
		}
	}
}

func event_arriveAtFloor(floor int, doorTimeoutTimer *elevalgo.ElevatorTimer, primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient){
	fmt.Println("client.event_arriveAtFloor:  Floor: ", floor)
	finRequest := elevalgo.Fsm_onFloorArrival(floor, doorTimeoutTimer)
	if(isConnected){
		for i := 0; i < len(finRequest); i++ {
			// Send completed request to primary
			completed_request := messages.M_CompletedRequest{Data:finRequest[i]}
			primary_socket.Out <- messages.MessageToBytes(completed_request)
			fmt.Println(" -> Sending comfirmed request to Primary nr:", i+1)
		}
	}
}

func event_primarySocketUpdate(primarySocketStatus bool, udp_socket *udpnetwork.ElevatorUDPClient, primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient) {
	if primarySocketStatus{
		connected := messages.M_Connected{}
		primary_socket.Out <- messages.MessageToBytes(connected)
		sendAndDeleteCabRequests(primary_socket)
		fmt.Println("client.event_primarySocketUpdate: Connection Succesfull")
		isConnected = true
	}else{
		isConnected = false
		fmt.Println("client.event_primarySocketUpdate: Retrying connection")
		go InitAndConnectToPrimaryServer(primary_socket, udp_socket)
	}
}

func event_doorTimeout(doorTimeoutTimer *elevalgo.ElevatorTimer, primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient) {
	fmt.Println("client.event_doorTimeout")
	finRequest := elevalgo.Fsm_onDoorTimeout(doorTimeoutTimer)
	if (isConnected){
		for i := 0; i < len(finRequest); i++ {
			// Send completed request to primary
			completed_request := messages.M_CompletedRequest{Data:finRequest[i]}
			primary_socket.Out <- messages.MessageToBytes(completed_request)
			fmt.Println(" -> Sending comfirmed request to Primary nr:", i+1)
		}
	}
}

func event_onObstruction(obstr bool){
	fmt.Println("client.event_onObstruction: ", obstr)
	elevalgo.Fsm_onObstruction(obstr)
}