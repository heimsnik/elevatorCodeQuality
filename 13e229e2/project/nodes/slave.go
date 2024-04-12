package nodes

import (
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"time"
	"../config"
	"../network"
	"../datatypes"
)


func startSlave() {
	
	listener, err := net.Listen("tcp", slavePort)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Slave node listening on port", slavePort)

	
	go connectToMaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleSlaveConnection(conn)
	}
}

func connectToMaster() {
	for {
		conn, err := net.Dial("tcp", "localhost"+masterPort)
		if err != nil {
			fmt.Println("Error connecting to master:", err)
			time.Sleep(5 * time.Second) // Retry 
			continue
		}

		go handleMasterCommunication(conn)
		return
	}
}

func handleMasterCommunication(conn net.Conn) {
	
	defer conn.Close()

	message := "Hello from slave"
	conn.Write([]byte(message))

	
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from master:", err)
			break
		}

		message := string(buffer[:n])
		fmt.Println("Received message from master:", message)

	
	}
}