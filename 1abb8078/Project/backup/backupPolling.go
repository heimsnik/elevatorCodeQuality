package backup

import (
	"Project/network/tcpnetwork"
	"time"
)

func pollPrimaryAlive(primary_socket *tcpnetwork.BackupToPrimaryTCPClient, primary_dead chan bool) {
	for {
		time.Sleep(_pollPrimaryAlive)
		if !primary_socket.IsActive(){
			primary_dead <- true
			return 
		}
	}
}

func periodicallySendAlive(primary_socket *tcpnetwork.BackupToPrimaryTCPClient, send_alive chan bool) {
	for {
		time.Sleep(_pollSendAlive)
		if !primary_socket.IsActive(){
			continue
		}

		send_alive <- true
	}
}

