package main

import (
	"Heisprosjekt/FSM"
	"Heisprosjekt/backup"
	"Heisprosjekt/driver-go-master/elevio"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/hra"
	"Heisprosjekt/network/bcast"
	"Heisprosjekt/network/peers"
	"Heisprosjekt/primary"
	"Heisprosjekt/runelevator"
	"Heisprosjekt/tcp"
	"Heisprosjekt/udp"
	"Heisprosjekt/utils"
)

func main() {

	elevio.Init("localhost:22343", elevcons.N_Floors)
	FSM.FSM_Initialize()

	drv_buttons := make(chan elevcons.ButtonEvent, 10)
	drv_floors := make(chan int, 10)
	drv_obstr := make(chan bool, 10)
	drv_stop := make(chan bool, 10)
	tcpReceiver := make(chan string, 10)
	udpReceiver := make(chan udp.Message, 10)
	udpTransmitter := make(chan udp.Message, 10)
	peerEnableCh := make(chan bool, 10)
	updateCh := make(chan peers.PeerUpdate, 10)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go primary.Primary_HallRequestSender(tcpReceiver)
	go backup.Backup_HallRequestSender()
	go tcp.TCP_MessageReader(utils.MyIP+elevcons.TcpPort, tcpReceiver)
	go udp.UDP_MessageSender(udpTransmitter)
	go hra.HRA_UpdateHRAElevstates()
	go bcast.Receiver(elevcons.MsgPort, udpReceiver)
	go bcast.Transmitter(elevcons.MsgPort, udpTransmitter)
	go peers.Transmitter(elevcons.UpdatePort, utils.MyIP, peerEnableCh)
	go peers.Receiver(elevcons.UpdatePort, updateCh)
	go udp.UDP_NetworkConnHandler(updateCh, udpReceiver, tcpReceiver, drv_buttons)
	go runelevator.RunElevator(tcpReceiver, drv_buttons, drv_floors, drv_obstr, drv_stop, peerEnableCh)

	select {}
}
