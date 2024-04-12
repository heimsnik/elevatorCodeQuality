package network

import (
	"Driver-go/config"
	"Driver-go/datatypes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

///////////// UDP

func UDP_broadcast() {
	udpAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255"+config.UdpMasterPort)
	if err != nil {
		fmt.Println("Error with UDP address: ", err.Error())
	}
	connUdp, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error while connecting: ", err.Error())
	}
	defer connUdp.Close()
	for {
		connUdp.Write([]byte("Hello from server"))
		time.Sleep(config.BroadcastInterval)
	}
}

func Find_server_ip() (bool, string) {
	masterExist := false
	var masterAddr string

	udpAddr, err := net.ResolveUDPAddr("udp", config.UdpMasterPort)
	if err != nil {
		fmt.Println("Error with UDP address: ", err.Error())
		return false, ""
	}

	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error while listning: ", err.Error())
		return false, ""
	}
	ln.SetDeadline(time.Now().Add(time.Second))
	defer ln.Close()

	buf := make([]byte, 1024)
	for !masterExist {
		n, addr, _ := ln.ReadFromUDP(buf)

		if string(buf[:n]) == "Hello from server" && addr != nil {
			masterExist = true
			masterAddr = addr.IP.String()
			fmt.Println("Found Master with address: ", masterAddr)
			break
		} else {
			break
		}
	}
	return masterExist, masterAddr
}

///////////// TCP

func Start_server(connLoss chan net.Addr, newConn chan net.Conn, newMsg chan datatypes.ServerMessage) {
	fmt.Print("Staring TCP server \n")
	tcpAddr, err := net.ResolveTCPAddr("tcp", config.TcpPort)
	if err != nil {
		fmt.Println("Error with TCP address: ", err.Error())
		return
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Error while listning: ", err.Error())
		return
	}
	for {
		connTcp, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println("Error while accepting: ", err.Error())
			break
		}
		newConn <- connTcp
		go Handle_message(connTcp, connLoss, newMsg)
	}
}

func Handle_message(conn *net.TCPConn, connLoss chan net.Addr, newMsg chan datatypes.ServerMessage) {
	msgAddr := conn.RemoteAddr()
	for {

		var msg datatypes.ServerMessage
		decoder := gob.NewDecoder(conn)
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("Lost connection to: ", msgAddr)
			connLoss <- msgAddr
			conn.Close()
			return
		}
		fmt.Println("Recieved msg from client with address: ", msgAddr)
		msg.Addr = msgAddr
		newMsg <- msg
	}
}

func Send_message(msg datatypes.ServerMessage, conn net.Conn) {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(msg)
	if err != nil {
		fmt.Println("Error sending: ", err.Error())
		return
	}
}

func Connect_to_server(masterAddr, client string) net.Conn {
	addr := masterAddr + config.TcpPort
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Print("Error with TCP address: ", err.Error())
	}
	connTcp, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Print("Error while connecting: ", err.Error())
	}
	fmt.Print("Connected to server\n")
	msg := datatypes.ServerMessage{Tag: "info", ClientInfo: client}
	Send_message(msg, connTcp)
	return connTcp
}

func Listen_to_server(receiver chan<- datatypes.ServerMessage, connLoss chan<- bool, conn net.Conn) {
	for {
		decoder := gob.NewDecoder(conn)
		var received datatypes.ServerMessage
		err := decoder.Decode(&received)
		if err != nil {
			fmt.Println("Error decoding/connection:", err.Error())
			connLoss <- true
			conn.Close()
			break
		}
		if received.Tag == "newTask" || received.Tag == "sync" {
			receiver <- received
		}
	}
}

