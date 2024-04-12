package udpnetwork

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type PrimaryUDPServer struct {
	BasicUDPSocket

	serverIp string
}

func NewPrimaryUDPServer() *PrimaryUDPServer {
	return &PrimaryUDPServer{
		BasicUDPSocket: *newBasicUDPSocket(),
		serverIp:       GetServerIP(),
	}
}

func (server *PrimaryUDPServer) Stop() {
	server.stopBasicSocket()
}
func (server *PrimaryUDPServer) BroadcastIP() {

	go func() {

		primaryOnlyPortInt, _ := strconv.Atoi(PRIMARY_ONLY_BROADCAST_PORT)
		primaryOnlyBroadcastConn := DialBroadcastUDP(primaryOnlyPortInt)
		defer primaryOnlyBroadcastConn.Close()
		primaryOnlyBroadcastAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", UDP_BROADCAST_IP, primaryOnlyPortInt))

		broadcastPortInt, _ := strconv.Atoi(UDP_BROADCAST_PORT)
		broadcastConn := DialBroadcastUDP(broadcastPortInt)
		defer broadcastConn.Close()
		broadcastAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", UDP_BROADCAST_IP, broadcastPortInt))

		for {
			time.Sleep(_broadcastInterval)
			select {
			case <-server.stop:
				return
			default:
				_, err := broadcastConn.WriteTo([]byte(server.serverIp), broadcastAddr)
				if err != nil {
					return
				}
				_, err = primaryOnlyBroadcastConn.WriteTo([]byte(server.serverIp), primaryOnlyBroadcastAddr)
				if err != nil {
					return
				}
			}
		}
	}()

}

func GetServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			if ipNet.IP.String()[:len(_labIPPrefix)] == _labIPPrefix {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

func (server *PrimaryUDPServer) CheckNetworkReachability() {
	go func() {
		addr, _ := net.ResolveUDPAddr("udp4", _randomIPandPort)
		_, err := net.DialUDP("udp4", nil, addr)

		if err != nil {
			server.In <- "No network"
			return
		}
		server.In <- "Network reachable"
	}()

}
