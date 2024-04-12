package tcp

import (
	"net"
	"time"
)

const _pollRate = 20 * time.Millisecond

func TCP_MakeConnection(BackupAddress string) *net.TCPConn {

	BackupTCPAdress, _ := net.ResolveTCPAddr("tcp4", BackupAddress)
	conn, _ := net.DialTCP("tcp4", nil, BackupTCPAdress)
	return conn
}

func TCP_MessageSender(receiverAddress string, HallRequests string) {

	conn := TCP_MakeConnection(receiverAddress)
	conn.Write([]byte(HallRequests))
	conn.Close()
}

func TCP_MessageReader(receiverAddress string, receiver chan<- string) {

	receiverTCPAddress, err := net.ResolveTCPAddr("tcp4", receiverAddress)
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 1024)
	for {
		listener, err := net.ListenTCP("tcp4", receiverTCPAddress)
		if err != nil {
			panic(err)
		}

		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		time.Sleep(_pollRate)

		n, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}

		if n != 0 {
			receiver <- string(buffer[:n])

		}
		conn.Close()
		listener.Close()
	}

}
