package tcpnetwork

import (
	"Project/network/messages"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PrimaryToElevatorTCPServer struct {
	BasicTCPSocket

	connManager *ConnectionManager
}

type ConnectionManager struct {
	connections map[int]Connection
	mutex       sync.RWMutex
}

type Connection struct {
	in        chan []byte
	out       chan []byte
	tcpSocket net.Conn
	active    bool
}

func NewPrimaryToElevatorTCPServer() *PrimaryToElevatorTCPServer {
	return &PrimaryToElevatorTCPServer{
		BasicTCPSocket: NewBasicTCPSocket(),
		connManager:    newConnectionManager(),
	}
}

func (server *PrimaryToElevatorTCPServer) Run() {
	go server.listenForNewElevators()
	go server.checkIfAlive()
	go server.routeDataToWrite()
}

func (server *PrimaryToElevatorTCPServer) Stop() {
	server.stopBasicSocket()
	activeConnections := server.GetActiveConnections()
	for _, ID := range activeConnections {
		server.deleteSingleConnection(ID)
	}
}

func (server *PrimaryToElevatorTCPServer) GetActiveConnections() []int {
	activeConnections := make([]int, 0)
	for ID, conn := range server.connManager.connections {
		if conn.active {
			activeConnections = append(activeConnections, ID)
		}
	}
	return activeConnections
}

func (server *PrimaryToElevatorTCPServer) IsActive() bool {
	return server.active
}

func (server *PrimaryToElevatorTCPServer) checkIfAlive() { 
	for {
		select {
		case <-server.stop:
			server.active = false
			return
		default:
			time.Sleep(_checkSocketAliveInterval)
			if len(server.GetActiveConnections()) > 0 {
				server.active = true
			} else {
				server.active = false
			}
		}
	}
}

func (server *PrimaryToElevatorTCPServer) deleteSingleConnection(ID int) { 
	_, ok := server.connManager.connections[ID]
	if ok {
		delete(server.connManager.connections, ID)
	}
}

func (server *PrimaryToElevatorTCPServer) routeDataToWrite() {
	for {
		select {
		case <-server.stop:
			return
		default:
			data_out := <-server.Out
			ID := messages.GetID(data_out)
			server.routeDatawMutex(ID, data_out)
		}
	}
}

func (server *PrimaryToElevatorTCPServer) routeDatawMutex(ID int, data_out []byte) {
	server.connManager.mutex.Lock()
	defer server.connManager.mutex.Unlock()

	if ID == 255 {
		for _, conn := range server.GetActiveConnections() {
			server.connManager.connections[conn].out <- data_out
		}

	} else {
		conn, ok := server.connManager.connections[ID]
		if ok && conn.active {
			conn.out <- data_out
		}
	}
}

func (server *PrimaryToElevatorTCPServer) listenForNewElevators() {
	newConnIDs := make(chan string)

	go server.acceptNewConnections(newConnIDs)

	for {
		select {
		case <-server.stop:
			return
		case new_port := <-newConnIDs:
			go server.connectToElevator(new_port)
		}

	}
}

func (server *PrimaryToElevatorTCPServer) acceptNewConnections(receiver chan<- string) {
	for {
		select {
		case <-server.stop:
			return
		default:
			ln, err := net.Listen("tcp", ":"+NEW_ELEVATOR_CONNECTION_PORT)
			if err != nil {
				return
			}
			defer ln.Close()

			ln.(*net.TCPListener).SetDeadline(time.Now().Add(_primaryAcceptConnectionTimeout))
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
			}else{
				server.giveElevatorPort(conn, receiver)
			}
			ln.Close()
		}

	}
}

func (server *PrimaryToElevatorTCPServer) giveElevatorPort(conn net.Conn, receiver chan<- string) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	ID := extractLastIpOctet(clientAddr)
	IDInt, _ := strconv.Atoi(ID)
	port := strconv.Itoa(ELEVATOR_PORT_OFFSET + IDInt)

	data := []byte("PORT:" + port)

	_, err := conn.Write(data)
	if err != nil {
		return
	}
	receiver <- ID
}

func extractLastIpOctet(clientAddr string) string {
	clientIP := strings.Split(clientAddr, ":")[0]
	octets := strings.Split(clientIP, ".")
	lastOctet := octets[len(octets)-1]
	return lastOctet
}

func (server *PrimaryToElevatorTCPServer) connectToElevator(ID string) {
	IDInt, _ := strconv.Atoi(ID)
	port := strconv.Itoa(ELEVATOR_PORT_OFFSET + IDInt)

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	ln.Close()

	server.connManager.addConnection(IDInt, Connection{in: make(chan []byte), out: make(chan []byte), tcpSocket: conn, active: true})

	go server.writeToElevator(IDInt)
	go server.readFromElevator(IDInt)
}

func (server *PrimaryToElevatorTCPServer) readFromElevator(ID int) {
	conn, ok := server.connManager.getConnection(ID)
	if !ok {
		return
	}

	buffer := make([]byte, _bufferSize)
	for {
		conn.tcpSocket.SetReadDeadline(time.Now().Add(IM_ALIVE_SIGNAL_MS_TIMEOUT))
		n, err := conn.tcpSocket.Read(buffer)

		if err != nil {
			server.deleteSingleConnection(ID)

			// Sleep to allow time for a slower localhost disconnect
			time.Sleep(_afterDeleteConnectionSleepTime)
			return
		}

		messageArray := messages.SplitMessages(buffer[:n])
		for _, segment := range messageArray {
			seg_len := len(segment)
			select{
			case server.In <- messages.SetID([]byte(segment), seg_len, ID):

			case <-server.stop:
				server.deleteSingleConnection(ID)
				return
			}
		}
	}
}

func (server *PrimaryToElevatorTCPServer) writeToElevator(ID int) {
	conn, ok := server.connManager.getConnection(ID)
	if !ok {
		return
	}

	for {
		select {
		case <-server.stop:
			return
		default:
			data_out := <-conn.out
			_, err := conn.tcpSocket.Write(data_out)
			if err != nil {
				return
			}
		}
	}
}

func newConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[int]Connection),
	}
}

func (cm *ConnectionManager) addConnection(ID int, conn Connection) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.connections[ID] = conn
}

func (cm *ConnectionManager) getConnection(ID int) (*Connection, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	conn, ok := cm.connections[ID]
	return &conn, ok
}
