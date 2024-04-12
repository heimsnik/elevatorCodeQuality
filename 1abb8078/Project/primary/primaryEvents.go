package primary

import (
	"Project/elevio"
	"Project/network/messages"
	"Project/network/tcpnetwork"
	"Project/network/udpnetwork"
	"fmt"
	"time"
)

func decodeElavatorMessage(bytes []byte,
	connected chan elevatorConnectedWithId,
	new_request chan requestWithId,
	completed_request chan requestWithId,
	elevator_alive chan elevatorAliveWithId) {

	messageArray := messages.SplitMessages(bytes)
	for _, segment := range messageArray {

		message := messages.BytesToMessage([]byte(segment))

		switch message.(type) {
		case messages.M_Connected:
			id := messages.GetID(bytes)
			connected <- elevatorConnectedWithId{id}
		case messages.M_NewRequest:
			id := messages.GetID(bytes)
			new_request <- requestWithId{id, message.(messages.M_NewRequest).Data}
		case messages.M_CompletedRequest:
			id := messages.GetID(bytes)
			btnEvent := message.(messages.M_CompletedRequest).Data
			completed_request <- requestWithId{id, btnEvent}
		case messages.M_ElevatorAlive:
			id := messages.GetID(bytes)
			elevator_alive <- elevatorAliveWithId{id: id, elevator: message.(messages.M_ElevatorAlive).Data}
		default:
			fmt.Println("primary.decodeElavatorMessage: Unknown message type:", segment)
		}
	}
}

func event_kill(elevator_server *tcpnetwork.PrimaryToElevatorTCPServer, backup_server *tcpnetwork.PrimaryToBackupTCPServer, udp_socket *udpnetwork.PrimaryUDPServer){
	elevator_server.Stop()
	backup_server.Stop()
	udp_socket.Stop()
	fmt.Println("Killing Primary")
}

func decodeBackupMessage(bytes []byte,
	c chan messages.M_Connected,
	abhr chan messages.M_AckBackupHallRequest,
	abcr chan messages.M_AckBackupCabRequest) {

	messageArray := messages.SplitMessages(bytes)
	for _, segment := range messageArray {

		message := messages.BytesToMessage([]byte(segment))

		switch message.(type) {
		case messages.M_Connected:
			c <- message.(messages.M_Connected)
		case messages.M_AckBackupHallRequest:
			abhr <- message.(messages.M_AckBackupHallRequest)
		case messages.M_AckBackupCabRequest:
			abcr <- message.(messages.M_AckBackupCabRequest)
		case messages.M_BackupAlive:
		default:
			fmt.Println("primary.decodeBackupMessage: Unknown message type:", segment)
		}
	}
}

func event_connectedElevator(a elevatorConnectedWithId, elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer) {
	fmt.Println("primary.event_connectedElevator: id:", a.id)

	connected_id := a.id
	_, exist := cabRequests[connected_id]
	
	// If elevator has been connected before
	if exist {
		req := cabRequests[connected_id]
		for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
			if req[floor] {
				time.Sleep(_maxSendRate)
				cabRequest := messages.M_DoRequest{Id: connected_id, Data: elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}}
				elevator_socket.Out <- messages.MessageToBytes(cabRequest)
				fmt.Println("primary.event_connectedElevator: Gave cab request to elevator floor:", floor)
			}

		}
	}
}

func event_connectedBackup(backup_socket *tcpnetwork.PrimaryToBackupTCPServer) {
	fmt.Println("primary.event_connectedBackup: Received connected backup message")

	onlyThisMachine = false

	// Send cab requests
	for id, reqs := range cabRequests {
		for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
			time.Sleep(_maxSendRate)
			if reqs[floor] {
				cabRequest := messages.M_BackupCabRequest{Id: id, Data: elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}}
				backup_socket.Out <- messages.MessageToBytes(cabRequest)
			} else {
				deleteCabRequest := messages.M_DeleteCabRequest{Id: id, Data: elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}}
				backup_socket.Out <- messages.MessageToBytes(deleteCabRequest)
			}
		}
	}

	// Send hall requests
	// Backup will ack, but we will not distribute again if the hall request is already active
	for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
		for btn := 0; btn < elevio.NUM_BUTTONS-1; btn++ {
			if hallRequests[floor][btn].active {
				hallRequest := messages.M_BackupHallRequest{Data: elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}}
				backup_socket.Out <- messages.MessageToBytes(hallRequest)
			} else {
				deleteHallRequest := messages.M_DeleteHallRequest{Data: elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}}
				backup_socket.Out <- messages.MessageToBytes(deleteHallRequest)
			}
		}
	}
}

func event_elevatorAlive(elevatorAlive elevatorAliveWithId) {
	// Update elevators map
	id := elevatorAlive.id
	elevators[id] = elevatorAlive.elevator
}

func event_sendAlive(elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer,
	backup_socket *tcpnetwork.PrimaryToBackupTCPServer) {
	// Send alive message to backup
	alive := messages.M_PrimaryAlive{Data: elevators}
	if backup_socket.IsActive() {

		backup_socket.Out <- messages.MessageToBytes(alive)
	}

	// Send alive message to elevators
	hallLights := messages.M_HallLights{Data: hallLightMatrix}
	elevator_socket.Out <- messages.MessageToBytes(hallLights) // will be broadcasted automatically
}

func event_newRequest(request requestWithId,
	backup_socket *tcpnetwork.PrimaryToBackupTCPServer,
	distribute_hallRequest chan messages.M_AckBackupHallRequest,
	give_cabRequest chan messages.M_AckBackupCabRequest) {

	fmt.Println("primary.event_newRequest: Received new request")
	var btnEvent = request.btnEvent
	var button = btnEvent.Button

	switch button {
	case elevio.BT_HallUp, elevio.BT_HallDown:
		// Distribute directly
		if onlyThisMachine {
			request := messages.M_AckBackupHallRequest{Data: btnEvent}
			distribute_hallRequest <- request
			return
		}
		// Send to backup
		request := messages.M_BackupHallRequest{Data: btnEvent}
		backup_socket.Out <- messages.MessageToBytes(request)
		fmt.Println("primary.event_newRequest: Sent hall request to backup", request)
	case elevio.BT_Cab:
		// Give cabRequest directly
		if onlyThisMachine {
			id := request.id
			request := messages.M_AckBackupCabRequest{Id: id, Data: btnEvent}
			give_cabRequest <- request
			return
		}
		// Send to backup
		id := request.id
		request := messages.M_BackupCabRequest{Id: id, Data: btnEvent}
		backup_socket.Out <- messages.MessageToBytes(request)
		fmt.Println("primary.event_newRequest: Sent cab request to backup", request)
	}
}

func event_completedRequest(request requestWithId,
	backup_socket *tcpnetwork.PrimaryToBackupTCPServer) {
	fmt.Println("primary.event_completedRequest: Received completed request")

	var btnEvent = request.btnEvent
	var floor = btnEvent.Floor
	var button = btnEvent.Button

	switch button {
	case elevio.BT_HallUp, elevio.BT_HallDown:
		// delete
		hallRequests[floor][button].active = false
		updateHallLightMatrix()

		if onlyThisMachine {
			return
		}

		// send delete to backup
		deleteHallRequest := messages.M_DeleteHallRequest{Data: btnEvent}
		backup_socket.Out <- messages.MessageToBytes(deleteHallRequest)
	case elevio.BT_Cab:
		// delete
		id := request.id
		reqs := cabRequests[id]
		reqs[floor] = false
		cabRequests[id] = reqs

		if onlyThisMachine {
			return
		}

		// send delete to backup
		deleteCabRequest := messages.M_DeleteCabRequest{Id: id, Data: btnEvent}
		backup_socket.Out <- messages.MessageToBytes(deleteCabRequest)
	}
}

func event_distributeHallRequest(request messages.M_AckBackupHallRequest,
	elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer) {

	var btnEvent = request.Data
	var floor = btnEvent.Floor
	var button = btnEvent.Button

	// ignore ack's on repeted requests from backup
	if hallRequests[floor][button].active {
		return
	}

	// Save request and start timer
	availableElevators := availableElevators(elevator_socket)
	id := elevatorToHandle(btnEvent, availableElevators)
	hallRequests[floor][button].active = true
	hallRequests[floor][button].id = id
	hallRequests[floor][button].timeAdded = time.Now()
	updateHallLightMatrix()

	// Distribute hallRequest
	hallRequest := messages.M_DoRequest{Id: id, Data: btnEvent}
	elevator_socket.Out <- messages.MessageToBytes(hallRequest)
	fmt.Println("primary.event_distributeHallRequest: Distributed request to elevator", id)
}

func event_giveCabRequest(request messages.M_AckBackupCabRequest,
	elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer) {

	var btnEvent = request.Data
	var id = request.Id
	var floor = btnEvent.Floor

	reqs := cabRequests[id]

	// save
	reqs[floor] = true
	cabRequests[id] = reqs

	// give cabRequest
	cabRequest := messages.M_DoRequest{Id: id, Data: btnEvent}
	elevator_socket.Out <- messages.MessageToBytes(cabRequest)
	fmt.Println("primary.event_giveCabRequest: Gave cab request to elevator", id)
}

func event_timeoutHallRequest(timedOutHallRequest timedOutHallRequest, elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer) {

	availableElevators := availableElevators(elevator_socket)
	
	// in no elevators available, only reset timers 
	if len(availableElevators) == 0 {
		hallRequests[timedOutHallRequest.floor][timedOutHallRequest.btn].timeAdded = time.Now()
		hallRequests[timedOutHallRequest.floor][timedOutHallRequest.btn].id = -1
		hallRequests[timedOutHallRequest.floor][timedOutHallRequest.btn].active = true
		fmt.Println("primary.event_timeoutHallRequest: No elevators available, reset timer")
		return
	}

	// find who to redistribute to (exclude the one who timed out if id != -1 and there is other to choose from)
	failed_id := timedOutHallRequest.id
	if failed_id != -1 && len(availableElevators) != 1 {
		for i := 0; i < len(availableElevators); i++ {
			if availableElevators[i] == failed_id {
				availableElevators = append(availableElevators[:i], availableElevators[i+1:]...)
				break
			}
		}

	}

	btnEvent := elevio.ButtonEvent{Floor: timedOutHallRequest.floor, Button: elevio.ButtonType(timedOutHallRequest.btn)}
	id_toHandle := elevatorToHandle(btnEvent, availableElevators)

	// save request and restart timer
	hallRequests[timedOutHallRequest.floor][timedOutHallRequest.btn].id = id_toHandle
	hallRequests[timedOutHallRequest.floor][timedOutHallRequest.btn].timeAdded = time.Now()
	hallRequests[timedOutHallRequest.floor][timedOutHallRequest.btn].active = true // should already be true
	updateHallLightMatrix()

	// send request
	hallRequest := messages.M_DoRequest{Id: id_toHandle, Data: btnEvent}
	elevator_socket.Out <- messages.MessageToBytes(hallRequest)
	fmt.Println("primary.event_timeoutHallRequest: Redistributed request to elevator", id_toHandle)
}

func event_backupDead(elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer,
	backup_socket *tcpnetwork.PrimaryToBackupTCPServer,
	machineID int) {
	fmt.Println("primary.event_backupDead: Backup is dead!")

	backup_socket.Run()

	// if no other machine, do nothing
	ids := elevator_socket.GetActiveConnections()
	if len(ids) == 1 && ids[0] == machineID {
		onlyThisMachine = true
		return
	}

	if time.Since(lastTimeSpawnBackupSent) > _minTimeBetweenSpawnBackup {
		// spawn backup on another machine
		for i := 0; i < len(ids); i++ {
			if ids[i] != machineID {
				spawnBackup := messages.M_SpawnBackup{Id: ids[i]}
				elevator_socket.Out <- messages.MessageToBytes(spawnBackup)
				lastTimeSpawnBackupSent = time.Now()
				fmt.Println("primary.event_backupDead: Sent spawn backup to machine", ids[i])
				return
			}
		}
	}
}
