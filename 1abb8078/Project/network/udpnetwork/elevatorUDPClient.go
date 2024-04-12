package udpnetwork

import (
	"net"
	"time"
)

type ElevatorUDPClient struct {
	BasicUDPSocket
}
func NewElevatorUDPClient() *ElevatorUDPClient {
	return &ElevatorUDPClient{
		BasicUDPSocket: *newBasicUDPSocket(),
	}
}

func (client *ElevatorUDPClient) Stop() {
	client.stopBasicSocket()
}

func (client *ElevatorUDPClient) ListenForUDPBroadcastedIP(broadcastPort string, stop <- chan bool) {
	go func() {

		conn, err := net.ListenPacket("udp", ":"+broadcastPort)
		if err != nil {
			return
		}
		defer conn.Close()

		buffer := make([]byte, _bufferSize)
		for{
			select{
				case <- client.stop:
					return
				case <- stop:
					return
				default:
					conn.SetReadDeadline(time.Now().Add(_udpReadTimeout))
					n, _, err := conn.ReadFrom(buffer)
					if err != nil {
						continue
					}
					select{
						case <- client.stop:
							return
						case <- stop:
							return
						case client.In <- string(buffer[:n]):
						default:
							buffer = make([]byte, _bufferSize)
				}
			}
		}
	}()
}
