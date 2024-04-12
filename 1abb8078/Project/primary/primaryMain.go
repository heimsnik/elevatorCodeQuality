package primary

import (
	"Project/elevalgo"
	"Project/elevio"
	"Project/network/messages"
	"Project/network/tcpnetwork"
	"Project/network/udpnetwork"
	"fmt"
	"time"
)

func PrimaryMain(oldElevators map[int]elevalgo.Elevator,
	confirmedHallRequests [elevio.NUM_FLOORS][elevio.NUM_BUTTONS - 1]bool,
	confirmedCabRequests map[int][elevio.NUM_FLOORS]bool) {

	fmt.Println("Primary - Initializing data")

	elevators = oldElevators
	hallRequests = resetAllTimesAndIDs(confirmedHallRequests)
	cabRequests = confirmedCabRequests
	lastTimeSpawnBackupSent = time.Now().Add( - _minTimeBetweenSpawnBackup)
	updateHallLightMatrix()

	onlyThisMachine = true
	hasBeenConnectedOnce = false

	// Find machine ID, we use last IP octet as IP
	ip := udpnetwork.GetServerIP()
	machineID := getMachineID(ip)
	
	// Check if other primary is broadcasting
	killIfOtherPrimaryBroadcast()
	
	// UDP Broadcast Init
	udp_socket := udpnetwork.NewPrimaryUDPServer()

	// Quit if there is no network connection
	udp_socket.CheckNetworkReachability()
	data:=  <-udp_socket.In
	if data  != "Network reachable" {
		fmt.Println("PrimaryMain: Network not reachable: Killing thread")
		return
	}

	// Start Broadcast
	udp_socket.BroadcastIP()

	// Initialize TCP Servers
	elevator_server := tcpnetwork.NewPrimaryToElevatorTCPServer()
	elevator_server.Run()

	backup_server := tcpnetwork.NewPrimaryToBackupTCPServer()
	backup_server.Run()

	// Channels
	elevator_connected := make(chan elevatorConnectedWithId)
	backup_connected := make(chan messages.M_Connected)

	new_request := make(chan requestWithId)
	completed_request := make(chan requestWithId)
	elevator_alive := make(chan elevatorAliveWithId)

	distribute_hallRequest := make(chan messages.M_AckBackupHallRequest)
	give_cabRequest := make(chan messages.M_AckBackupCabRequest)

	timeout_hallRequest := make(chan timedOutHallRequest)

	backup_dead := make(chan bool)
	send_alive := make(chan bool)
	kill := make(chan bool)
	
	// Start periodical routines

	go pollBackupAlive(backup_server, backup_dead)
	go periodicallySendAlive(send_alive)
	go pollHallRequestTimeout(timeout_hallRequest)
	go pollPrimaryConnected(elevator_server, backup_dead,kill)

	for {
		select {
		case a := <-elevator_server.In:
			go decodeElavatorMessage(a, elevator_connected, new_request, completed_request, elevator_alive)
		case <-kill:
			event_kill(elevator_server, backup_server, udp_socket)
			return
		case a := <-backup_server.In:
			go decodeBackupMessage(a, backup_connected, distribute_hallRequest, give_cabRequest)

		case a := <-elevator_connected:
			event_connectedElevator(a, elevator_server)

		case <-backup_connected:
			event_connectedBackup(backup_server)

		case a := <-elevator_alive:
			event_elevatorAlive(a)

		case <-send_alive:
			event_sendAlive(elevator_server, backup_server)

		case a := <-new_request:
			go event_newRequest(a, backup_server, distribute_hallRequest, give_cabRequest)

		case a := <-completed_request:
			event_completedRequest(a, backup_server)

		case a := <-distribute_hallRequest:
			event_distributeHallRequest(a, elevator_server)

		case a := <-give_cabRequest:
			event_giveCabRequest(a, elevator_server)

		case a := <-timeout_hallRequest:
			event_timeoutHallRequest(a, elevator_server)

		case <-backup_dead:
			event_backupDead(elevator_server, backup_server,machineID)

		}
	}
}
