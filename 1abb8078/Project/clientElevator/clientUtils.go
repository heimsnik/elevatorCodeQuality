package clientElevator

import (
	"Project/elevalgo"
	"Project/elevio"
	"Project/network/messages"
	"Project/network/tcpnetwork"
	"Project/network/udpnetwork"
	"Project/primary"
	"fmt"
	"time"
)

func SpawnPrimaryIfAlone(udp_socket_In chan string){
	spawnPrimaryTimer := time.NewTimer(_primarySpawnTime)
	select {
	case <-spawnPrimaryTimer.C:
		fmt.Println("--Creating primary--")
		go primary.PrimaryMain(map[int]elevalgo.Elevator{}, [elevio.NUM_FLOORS][2]bool{}, make(map[int][elevio.NUM_FLOORS]bool))
	case <-udp_socket_In:
	}
}

func InitAndConnectToPrimaryServer(primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient, udp_socket *udpnetwork.ElevatorUDPClient){
	for{
		ip_primary := <- udp_socket.In
		primary_socket.SetPrimaryIP(ip_primary)
		primary_socket.Run()
		time.Sleep(_reattemptConnectToPrimaryTime)
		if (primary_socket.IsActive()){
			return
		}
	}
}

func sendAndDeleteCabRequests(primary_socket *tcpnetwork.ElevatorToPrimaryTCPClient){
	for i := 0; i < elevio.NUM_FLOORS; i++ {
		if(elevalgo.GetElevator().Requests[i][elevio.BT_Cab]) {
			cabRequest := messages.M_NewRequest{Data: elevio.ButtonEvent{Floor: i, Button: elevio.BT_Cab}}
			primary_socket.Out <- messages.MessageToBytes(cabRequest)
			elevalgo.Fsm_onReconnectClearCabRequest(i)
			fmt.Println("client.sendAndDeleteCabRequest -> Sent cab request (floor:", i ,") to Primary and deleted it")
		}
	}
}