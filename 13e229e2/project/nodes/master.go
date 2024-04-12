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

func SyncBackupQ(conn* net.Conn, []mstq datatypes.MasterQueue ) {
	// make sure the state of both master and it's backup are the same
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(mstq)
		if err != nil {
			fmt.Println("Error encoding:", err.Error())
			return
		}
	fmt.Println("Syncing backup")
}

func SyncBackupO(conn* net.Conn, overview datatypes.Overview  ) {
	// make sure the state of both master and it's backup are the same
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(overview)
		if err != nil {
			fmt.Println("Error encoding:", err.Error())
			return
		}
	fmt.Println("Syncing backup")
}

func updateMasterQueue(mstq MasterQueue, newtask datatypes.elevInfo) {
	mstq.mu.Lock() // mutex to avoid race condition
	append(mstq.Elevators, newtask)
	mstq.mu.Unlock()
}

func updateMasterOverview(currentOverview datatypes.OverviewList, newoverview datatypes.Overview) {
	currentOverview.mu.Lock()
	append(currentOverview.List, newoverview)
	currentOverview.mu.Unlock()
}



func startMaster() {

	ln, err := net.Listen("tcp", masterPort)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Master node listening on port", masterPort)

	// Start a backup node
	go spawnBackup()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleMasterConnection(conn)
	}
}


func handleMasterConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			break
		}

		message := string(buffer[:n])
		fmt.Println("Received message from slave:", message)
	}
}

func spawnBackupIfNeeded() {
	
	_, err := net.Dial("tcp", "localhost"+backupPort)
	if err != nil {
		go spawnBackup()
	}
}

func sendMessageToNode(nodePort string, message string) {
	conn, err := net.Dial("tcp", "localhost"+nodePort)
	if err != nil {
		fmt.Println("Error connecting to node:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(message))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading response from node:", err)
		return
	}

	response := string(buffer[:n])
	fmt.Println("Received response from node:", response)
}
