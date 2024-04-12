package clientElevator

import (
	"Project/elevalgo"
	"Project/elevio"
	"Project/network/messages"
	"Project/network/tcpnetwork"
	"Project/network/udpnetwork"
	"fmt"
)

func ClientMain() {

	fmt.Println("Client Init Started")

	// Init local elevator:
	elevio.Init("localhost:15657")

	// Channels:
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	message_do_request := make(chan messages.M_DoRequest)
	hall_lights := make(chan messages.M_HallLights)
	spawn_backup := make(chan messages.M_SpawnBackup)
	message_kill := make(chan messages.M_KILL)
	send_alive := make(chan bool)
	primary_socket_connect := make(chan bool)

	// Timers:
	doorTimeoutTimer := elevalgo.NewTimer()

	// Init network and start Primary on timeout
	udp_socket := udpnetwork.NewElevatorUDPClient()
	defer udp_socket.Stop()
	udp_socket.ListenForUDPBroadcastedIP(udpnetwork.UDP_BROADCAST_PORT, make(<-chan bool))

	// Spawn primary if alone
	SpawnPrimaryIfAlone(udp_socket.In)
	
	// TCP client to primary
	primary_socket := tcpnetwork.NewElevatorToPrimaryTCPClient("")
	defer primary_socket.Stop()

	// Connect to primary
	go InitAndConnectToPrimaryServer(primary_socket, udp_socket)

	// Polling
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go pollSendAlive(send_alive)
	go pollPrimaryConnection(primary_socket, primary_socket_connect)
	go pollLights()

	// Init Elevator in start state:
	elevalgo.Fsm_onInitBetweenFloors()

	fmt.Println("Elevator Init Done")
	for {
		select {
		case a := <-primary_socket.In:
			go decodeMessage(a, message_do_request, hall_lights, spawn_backup, message_kill)
		case a := <-message_do_request:
			go event_doRequest(a, doorTimeoutTimer, primary_socket)
		case a := <-hall_lights:
			event_doHallLights(a)
		case <-spawn_backup:
			go event_spawnBackup(udp_socket)
		case <-message_kill:
			event_doKill(primary_socket, udp_socket)
		case <-send_alive:
			go event_sendAlive(primary_socket)
		case a := <-primary_socket_connect:
			go event_primarySocketUpdate(a, udp_socket, primary_socket)
		case a := <-drv_buttons:
			go event_buttonPress(a, doorTimeoutTimer, primary_socket)
		case a := <-drv_floors:
			go event_arriveAtFloor(a, doorTimeoutTimer, primary_socket)
		case <-doorTimeoutTimer.Timer.C:
			go event_doorTimeout(doorTimeoutTimer, primary_socket)
		case a := <-drv_obstr:
			event_onObstruction(a)
		}
	}
}
