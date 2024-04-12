package primary

import (
	"Project/network/tcpnetwork"
	"Project/elevio"
	"time"
	"fmt"
)

func pollBackupAlive(backup_socket *tcpnetwork.PrimaryToBackupTCPServer, backup_dead chan bool) {
	for {
		time.Sleep(_pollRateBackup)
		if !backup_socket.IsActive() {
			backup_dead <- true
		}
	}
}

func periodicallySendAlive(send_alive chan bool) {
	for {
		time.Sleep(_pollRateSendAlive)
		send_alive <- true
	}
}

func pollPrimaryConnected(elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer, backup_dead chan bool, kill chan <- bool) {
	for {
		time.Sleep(_pollPrimaryAliveRate)

		if elevator_socket.IsActive(){
			hasBeenConnectedOnce = true
		}

		fails := 0
		for i := 0; i < _pollPrimaryAliveMaxFails ; i++ {
			time.Sleep(_pollPrimaryAliveSubtick)
			if elevator_socket.IsActive() || !hasBeenConnectedOnce{
				break
			}
			fmt.Println(elevator_socket.GetActiveConnections())
			fmt.Println(elevator_socket.IsActive())
			fails++
		}
		if fails == _pollPrimaryAliveMaxFails  {
			t := time.NewTimer(_pollRateBackup)
			select {
			case <-t.C:
				continue
			case <-backup_dead:
				kill <- true
				return
			}
		}
	}
}

func pollHallRequestTimeout(timeout_hallRequest chan timedOutHallRequest) {
	for {
		time.Sleep(_pollRateTimeout)

		for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
			for btn := 0; btn < elevio.NUM_BUTTONS-1; btn++ {
				if hallRequests[floor][btn].active {
					if time.Since(hallRequests[floor][btn].timeAdded) > _maxTimeOnRequest {
						fmt.Printf("Request at floor %d, button %d has been waiting for to long!\n", floor, btn)
						timeout_hallRequest <- timedOutHallRequest{id: hallRequests[floor][btn].id, floor: floor, btn: btn}
					}
				}
			}
		}
	}
}
