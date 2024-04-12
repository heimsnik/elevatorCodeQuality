package backup

import (
	"Project/elevio"
	"Project/network/messages"
	"Project/network/tcpnetwork"
	"fmt"
)

func decodeMessage(bytes []byte,
	bhr chan messages.M_BackupHallRequest, 
	bcr chan messages.M_BackupCabRequest,
	dhr chan messages.M_DeleteHallRequest,
	dcr chan messages.M_DeleteCabRequest,
	pa chan messages.M_PrimaryAlive) {
	
	messageArray := messages.SplitMessages(bytes)
	for _, segment := range messageArray{

		message := messages.BytesToMessage([]byte(segment))
		switch message.(type){
		case messages.M_BackupHallRequest:
			bhr <- message.(messages.M_BackupHallRequest)
		case messages.M_BackupCabRequest:
			bcr <- message.(messages.M_BackupCabRequest)
		case messages.M_DeleteHallRequest:
			dhr <- message.(messages.M_DeleteHallRequest)
		case messages.M_DeleteCabRequest:
			dcr <- message.(messages.M_DeleteCabRequest)
		case messages.M_PrimaryAlive:
			pa <- message.(messages.M_PrimaryAlive) 
		default:
			fmt.Println("backup.decodeMessage: Unknown message type : ", string(bytes))
		}
	}
}

func event_backupHallRequest(request messages.M_BackupHallRequest, primary_socket *tcpnetwork.BackupToPrimaryTCPClient) {
	// Save request
	fmt.Println("backup.event_backupRequest: Received backup request")
	var btnEvent elevio.ButtonEvent = request.Data
	var floor int = btnEvent.Floor
	var button elevio.ButtonType = btnEvent.Button
	confirmedHallRequests[floor][button] = true 

	// Ack to primary
	ack := messages.M_AckBackupHallRequest{Data: btnEvent}
	primary_socket.Out <- messages.MessageToBytes(ack)
}

func event_backupCabRequest(request messages.M_BackupCabRequest, primary_socket *tcpnetwork.BackupToPrimaryTCPClient) {
	// Save request
	fmt.Println("backup.event_backupRequest: Received backup request")
	var btnEvent elevio.ButtonEvent = request.Data
	var id int = request.Id
	var floor int = btnEvent.Floor
	reqs := confirmedCabRequests[id]
	reqs[floor] = true
	confirmedCabRequests[id] = reqs
	
	// Ack to primary
	ack := messages.M_AckBackupCabRequest{Id: id, Data: btnEvent}
	primary_socket.Out <- messages.MessageToBytes(ack)
}

func event_deleteHallRequest(request messages.M_DeleteHallRequest) {
	fmt.Println("backup.event_deleteHallRequest: Received delete hallRequest")
	var btnEvent elevio.ButtonEvent = request.Data
	var floor int = btnEvent.Floor
	var button elevio.ButtonType = btnEvent.Button
	confirmedHallRequests[floor][button] = false
}

func event_deleteCabRequest(request messages.M_DeleteCabRequest) {
	fmt.Println("backup.event_deleteRequest: Received delete cabRequest")
	var btnEvent elevio.ButtonEvent = request.Data
	var id int = request.Id
	var floor int = btnEvent.Floor
	reqs := confirmedCabRequests[id]
	reqs[floor] = false
	confirmedCabRequests[id] = reqs
}

func event_primaryAlive(e messages.M_PrimaryAlive) {
	elevators = e.Data
}

func event_sendAlive(primary_socket *tcpnetwork.BackupToPrimaryTCPClient) {
	alive := messages.M_BackupAlive{}
	primary_socket.Out <- messages.MessageToBytes(alive)
}