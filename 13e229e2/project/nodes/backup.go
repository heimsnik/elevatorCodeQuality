package nodes

import (
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"time"
	"encoding/gob"
	"../config"
	"../datatypes"
)


func spawnBackup() {
	
	ln, err := net.Listen("tcp", config.UdpBackupPort)
	if err != nil {
		fmt.Println("Error listening for backup:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Backup node listening on port", config.UdpBackupPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection for backup:", err)
			continue
		}
		go handleBackupConnection(conn)
	}
}



func SyncBackup(conn net.Conn) {
	defer conn.Close()
	
	decoder := gob.NewDecoder(conn)

	for {
		var recvOverview datatypes.Overview
		err := decoder.Decode(&recvOverview)
		if err != nil {
			fmt.Println("Error decoding message:", err)
			break
		}
		current_overview = recvOverview

	}
}




