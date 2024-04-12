package tcpnetwork

import (
	"net"
	"time"
)

type ElevatorToPrimaryTCPClient struct {
	BasicTCPSocket

	primaryIP string
}

func NewElevatorToPrimaryTCPClient(primaryIP string) *ElevatorToPrimaryTCPClient {
	return &ElevatorToPrimaryTCPClient{
		BasicTCPSocket: NewBasicTCPSocket(),
		primaryIP:      primaryIP,
	}
}

func (client *ElevatorToPrimaryTCPClient) Run() {
	go func() {
		receivedConnPort := make(chan string)
		go client.getPortFromPrimary(receivedConnPort)
		port := <-receivedConnPort
		go client.connectToPrimary(port)
	}()
}

func (client *ElevatorToPrimaryTCPClient) Stop() {
	client.active = false
	client.stopBasicSocket()
}

func (client *ElevatorToPrimaryTCPClient) SetPrimaryIP(ip string) {
	client.primaryIP = ip
}

func (client *ElevatorToPrimaryTCPClient) connectToPrimary(port string) {
	conn, err := net.Dial("tcp", client.primaryIP+":"+port)
	if err != nil {
		return
	}
	client.active = true

	go client.readFromPrimary(conn)
	go client.writeToPrimary(conn)

}

func (client *ElevatorToPrimaryTCPClient) writeToPrimary(conn net.Conn) {
	for{
		select{
			case <- client.stop:
				return
			default:
				data_out := <-client.Out
				_, err := conn.Write(data_out)
				if err != nil {
					return
				}
			}
	}
}

func (client *ElevatorToPrimaryTCPClient) readFromPrimary(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, _bufferSize)

	for {
		select{
			case <- client.stop:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(IM_ALIVE_SIGNAL_MS_TIMEOUT))
				n, err := conn.Read(buffer)
				if err != nil {
					client.active = false
					return
				}
				select{
					case client.In <- buffer[:n]:
					case <- client.stop:
						return
				}
				
		}
	}
}

func (client *ElevatorToPrimaryTCPClient) getPortFromPrimary(receiver chan<- string) { //Connects to primary and sends a message to the primary. This needs to be made into class functions
	conn, _ := net.Dial("tcp", client.primaryIP+":"+NEW_ELEVATOR_CONNECTION_PORT)
	defer conn.Close()


	buffer := make([]byte, _bufferSize)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		if string(buffer[:len("PORT:")]) == "PORT:" {
			receiver <- string(buffer[5:n])
			return
		}
	}
}
