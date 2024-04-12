package tcpnetwork

import (
	"fmt"
	"net"
	"time"
)


type BackupToPrimaryTCPClient struct {
	BasicTCPSocket

	primaryIP    string
}

func NewBackupToPrimaryTCPClient(primaryIP string) *BackupToPrimaryTCPClient {
	return &BackupToPrimaryTCPClient{
		BasicTCPSocket:	   NewBasicTCPSocket(),
		primaryIP:   	   primaryIP,
	}
}

func (client *BackupToPrimaryTCPClient) Run() {
	go client.connectToPrimary() 
}

func (client *BackupToPrimaryTCPClient) Stop() {
	client.stopBasicSocket()
}

func (client *BackupToPrimaryTCPClient) connectToPrimary() {
	conn, err := net.Dial("tcp", client.primaryIP+":"+BACKUP_PORT)
	if err != nil {
		return
	}
	fmt.Println("Backup connected to primary")
	client.active = true
	
	go client.readFromPrimary(conn)
	go client.writeToPrimary(conn)

}

func (client *BackupToPrimaryTCPClient) writeToPrimary(conn net.Conn) {
	for {
		select{
		case <-client.stop:
			return
		default:
			data_out := <-client.Out
			conn.SetWriteDeadline(time.Now().Add(_tcpWriteDeadline))
			_, err := conn.Write(data_out)
			if err != nil {
				return
			}
		}
	}
}

func (client *BackupToPrimaryTCPClient) readFromPrimary(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, _bufferSize)
	
	for {
		conn.SetReadDeadline(time.Now().Add(IM_ALIVE_SIGNAL_MS_TIMEOUT))
		n, err := conn.Read(buffer)
		if err != nil {
			client.active = false
			return
		}
		select{
		case <-client.stop:
			return
		case client.In <- buffer[:n]:
		}
		
	}

}