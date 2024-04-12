package udp

import (
	"Heisprosjekt/FSM"
	"Heisprosjekt/elevator"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/hra"
	"Heisprosjekt/network/peers"
	"Heisprosjekt/utils"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type Message struct {
	Elev     elevator.Elevator
	IP       string
	SenderIP string
}

func filterOutHallRequests(requests [4][3]int) [4][3]int {

	for floor, btnVec := range requests {
		for btn := range btnVec {
			if btn != int(elevcons.BT_Cab) {
				requests[floor][btn] = 0
			}
		}
	}
	return requests
}

func mergeLights(lights [4][3]int) [4][3]int {

	for floor, floorVec := range utils.MyElev.Lights {
		for btn, light := range floorVec {
			if btn != int(elevcons.BT_Cab) {
				lights[floor][btn] = light
			}
		}
	}
	return lights
}

func UDP_MessageSender(transmitter chan<- Message) {

	for {
		if len(utils.NeedInfoElevs) == 0 {
			transmitter <- Message{utils.MyElev, utils.MyIP, utils.MyIP}
			time.Sleep(time.Second * 1)

		} else {
			transmitter <- Message{utils.MyElev, utils.MyIP, utils.MyIP}
			time.Sleep(time.Second * 1)
			for ip, e := range utils.NeedInfoElevs {
				e.Lights = mergeLights(e.Lights)
				transmitter <- Message{e, ip, utils.MyIP}
				time.Sleep(time.Second * 1)
			}
		}
	}
}

func getMyElevatorMode(peers []string) elevcons.ElevatorMode {

	myID := utils.MyIP[len(utils.MyIP)-2:]
	allIDs := make([]int, 0)
	numPeers := len(peers)

	if numPeers == 0 {
		return elevcons.SingleElevator
	} else if numPeers == 1 && peers[0] == utils.MyIP {
		return elevcons.SingleElevator
	}

	for i := 0; i < numPeers; i++ {
		id, _ := strconv.Atoi(peers[i][len(peers[i])-2:])
		allIDs = append(allIDs, id)
	}

	sort.Ints(allIDs)
	newPrimaryID := fmt.Sprint(allIDs[0])

	if myID == newPrimaryID {
		return elevcons.Primary
	} else {
		return elevcons.Backup
	}
}

func mergeLostInfo(lostInfo elevator.Elevator, drv_buttons chan elevcons.ButtonEvent) {

	utils.MyElevsMutex.Lock()
	if utils.MyElev.ReturnStatus == elevcons.NeedAllInfo {
		for floor, floorReq := range lostInfo.Requests {
			for btn, req := range floorReq {
				if req == 1 {
					drv_buttons <- elevcons.ButtonEvent{Floor: floor, Button: elevcons.ButtonType(btn)}
					utils.MyElev.Requests[floor][btn] = 1
				}
			}
		}
		utils.MyElev.Lights = lostInfo.Lights
	} else if utils.MyElev.ReturnStatus == elevcons.NeedLightInfo {
		for floor, floorVec := range lostInfo.Lights {
			for btn, light := range floorVec {
				if btn != 2 {
					utils.MyElev.Lights[floor][btn] = light
				}
			}
		}
	}
	utils.MyElevsMutex.Unlock()
	FSM.FSM_SetAllLights()
}

func saveLostHallCalls(lostElevs map[string]elevator.Elevator) {

	for _, elev := range lostElevs {
		requests := elev.Requests
		for floor, floorReq := range requests {
			for btn, req := range floorReq {
				if req == 1 && btn != int(elevcons.BT_Cab) {
					utils.LostHallCalls = append(utils.LostHallCalls, [3]int{0, floor, btn})
				}
			}
		}
	}
}

func registerLostElevators(peerInfo peers.PeerUpdate, lostElevs *map[string]elevator.Elevator) {

	if len(peerInfo.Lost) == 0 {
		return
	}

	for _, ip := range peerInfo.Lost {
		(*lostElevs)[ip] = utils.CurrentElevs[ip]

		utils.CurrentElevsMutex.Lock()
		delete((utils.CurrentElevs), ip)
		utils.CurrentElevsMutex.Unlock()

		utils.CurrentElevsMutex.Lock()
		delete((hra.HRAElevstates), ip)
		utils.CurrentElevsMutex.Unlock()
	}

}

func registerReturnedElevators(peerInfo peers.PeerUpdate, lostElevs *map[string]elevator.Elevator) {

	_, lostElev := (*lostElevs)[peerInfo.New]

	if lostElev {

		utils.NeedInfoElevsMutex.Lock()
		utils.NeedInfoElevs[peerInfo.New] = (*lostElevs)[peerInfo.New]
		utils.NeedInfoElevsMutex.Unlock()

		delete(*lostElevs, peerInfo.New)
	}
}

func saveMyLostInfo(message Message, drv_buttons chan elevcons.ButtonEvent) {

	mergeLostInfo(message.Elev, drv_buttons)

	utils.MyElevsMutex.Lock()
	utils.MyElev.ReturnStatus = elevcons.NoInfoNeeded
	utils.MyElevsMutex.Unlock()
}

func updateTCPHallReqInstrMap(peers []string) {

	if utils.MyElev.Status == elevcons.Primary {

		utils.HallreqInstrMutex.Lock()
		for _, ip := range peers {
			utils.HallreqInstructions[ip] = [][3]int{}
		}
		utils.HallreqInstrMutex.Unlock()
	}
}

func UDP_NetworkConnHandler(peerUpdate chan peers.PeerUpdate, reciever chan Message, tcpReceiver chan string, drv_buttons chan elevcons.ButtonEvent) {

	lostElevs := make(map[string]elevator.Elevator)
	lostInfoTimer := time.NewTimer(3 * time.Second)

	<-peerUpdate
	singleElevTimer := time.NewTimer(500 * time.Millisecond)

	for {
		select {
		case <-singleElevTimer.C:

			utils.MyElevsMutex.Lock()
			utils.MyElev.Status = elevcons.SingleElevator
			utils.MyElevsMutex.Unlock()

		case <-lostInfoTimer.C:

			utils.MyElevsMutex.Lock()
			utils.MyElev.ReturnStatus = elevcons.NoInfoNeeded
			utils.MyElevsMutex.Unlock()

		case peerInfo := <-peerUpdate:
			singleElevTimer.Stop()

			registerLostElevators(peerInfo, &lostElevs)
			registerReturnedElevators(peerInfo, &lostElevs)

			utils.MyElevsMutex.Lock()
			utils.MyElev.Status = getMyElevatorMode(peerInfo.Peers)
			utils.MyElevsMutex.Unlock()

			updateTCPHallReqInstrMap(peerInfo.Peers)

			if utils.MyElev.Status == elevcons.Backup {
				break
			}

			if utils.MyElev.Status == elevcons.SingleElevator {
				utils.MyElevsMutex.Lock()
				utils.MyElev.ReturnStatus = elevcons.NeedLightInfo
				utils.MyElev.Requests = filterOutHallRequests(utils.MyElev.Requests)
				utils.MyElev.Lights = utils.MyElev.Requests
				utils.MyElevsMutex.Unlock()
				break
			}

			if utils.MyElev.Status == elevcons.Primary {
				utils.PrimatyIPMutex.Lock()
				utils.PrimaryIP = utils.MyIP
				utils.PrimatyIPMutex.Unlock()

				saveLostHallCalls(lostElevs)
				if len(utils.LostHallCalls) != 0 {
					tcpReceiver <- hra.HRA_RequestMatrixToString(utils.LostHallCalls)
					utils.LostHallCalls = [][3]int{}
				}
			}

		case message := <-reciever:

			utils.CurrentElevsMutex.Lock()
			utils.CurrentElevs[message.IP] = message.Elev
			utils.CurrentElevsMutex.Unlock()

			hra.HRA_UpdateHRAElevstates()

			if message.SenderIP == utils.MyIP {
				break
			}

			if message.IP == utils.MyIP {
				saveMyLostInfo(message, drv_buttons)
				break
			}

			if message.Elev.Status == elevcons.Primary {

				utils.PrimatyIPMutex.Lock()
				utils.PrimaryIP = message.IP
				utils.PrimatyIPMutex.Unlock()

			}

			_, inNeedInfoElevs := utils.NeedInfoElevs[message.IP]
			if inNeedInfoElevs && message.Elev.ReturnStatus == elevcons.NoInfoNeeded {
				utils.NeedInfoElevsMutex.Lock()
				delete(utils.NeedInfoElevs, message.IP)
				utils.NeedInfoElevsMutex.Unlock()
			}
		}
	}
}
