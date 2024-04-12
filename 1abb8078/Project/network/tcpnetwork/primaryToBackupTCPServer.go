package tcpnetwork

import (
	"net"
	"time"
)

type PrimaryToBackupTCPServer struct {
	BasicTCPSocket
}

func NewPrimaryToBackupTCPServer() *PrimaryToBackupTCPServer {
	return &PrimaryToBackupTCPServer{
		BasicTCPSocket: NewBasicTCPSocket(),
	}
}

func (server *PrimaryToBackupTCPServer) Stop() {
	server.stopBasicSocket()
}

func (server *PrimaryToBackupTCPServer) Run() {
	go server.connectToBackup()
}

func (server *PrimaryToBackupTCPServer) IsActive() bool {
	return server.active
}

func (server *PrimaryToBackupTCPServer) connectToBackup() {
	for{
		select{
		case <-server.stop:
			return
		default:
			ln, err := net.Listen("tcp", ":"+BACKUP_PORT)
			if err != nil {
				return
			}
			ln.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second))
			conn, err := ln.Accept()
			if err != nil {
				ln.Close()
				continue
			}

			server.active = true
			ln.Close()
			go server.readFromBackup(conn)
			go server.writeToBackup(conn)
			return
			}
		}
	}

func (server *PrimaryToBackupTCPServer) writeToBackup(conn net.Conn) {
	for {
		data_out := <-server.Out
		_, err := conn.Write(data_out)
		if err != nil {
			return
		}
	}
}

func (server *PrimaryToBackupTCPServer) readFromBackup(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, _bufferSize)
	for {
		select{
		case <-server.stop:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(IM_ALIVE_SIGNAL_MS_TIMEOUT))
			n, err := conn.Read(buffer)
			if err != nil {
				server.active = false
				return
			}
			server.In <- buffer[:n]
		}	
	}
}
