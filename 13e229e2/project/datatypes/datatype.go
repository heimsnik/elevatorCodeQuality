package datatypes

import (
	"Driver-go/elevio"
	"net"
	"sync"
	"Driver-go/config"
)

//////////// Datatype of messages between client and server

type ServerMessage struct {
	Tag              string					//Describes purpose of the message
	Task             elevio.ButtonEvent
	Floor      		 int
	ClientInfo       string
	Queue      		 []Queue
	Overview   		 []Overview
	Addr             net.Addr
}

//////////// Datatypes master

type Queue struct {
	Floor          int
	Button         elevio.ButtonType
	ReceivedFrom   net.Addr
	SendTo         net.Addr
}

type Overview struct {
	ClientType	 string		//backup or elevator
	Id           int
	CurrentFloor int
	ConnStatus   bool
	Task		 []elevio.ButtonEvent
	Addr         net.Addr
	Conn         net.Conn
}

type elevInfo struct {
	Floor 				int
	Button 				elevio.ButtonType
	ReceivedFromId 		int
	SendToId 			int
}

type MasterQueue struct {
	Elevators [config.NumElevators]elevInfo
	mu sync.Mutex
}

type OverviewList struct{
	List []Overview
	mu sync.Mutex
}
