package primary

import (
	"Heisprosjekt/FSM"
	"Heisprosjekt/driver-go-master/elevio"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/hra"
	"Heisprosjekt/tcp"
	"Heisprosjekt/utils"
	"time"
)

func Primary_HallRequestSender(tcpReceiver chan<- string) {

	for {
		time.Sleep(250 * time.Millisecond)

		if utils.MyElev.Status == elevcons.Primary {
			for IP, req := range utils.HallreqInstructions {
				if len(req) == 0 {
					break
				}
				_, amongCurrentElevs := utils.CurrentElevs[IP]
				if amongCurrentElevs {
					if IP != utils.MyIP {
						tcp.TCP_MessageSender(IP+elevcons.TcpPort, hra.HRA_RequestMatrixToString(req))
					} else {
						tcpReceiver <- hra.HRA_RequestMatrixToString(utils.HallreqInstructions[IP])
					}
				}

				utils.HallreqInstrMutex.Lock()
				utils.HallreqInstructions[IP] = [][3]int{}
				utils.HallreqInstrMutex.Unlock()

			}
		}
	}
}

func executeAssignedRequests(assignedRequests [][3]int) {

	for i := 0; i < len(assignedRequests); i++ {
		btn_floor := assignedRequests[i][1]
		btn_type := elevcons.ButtonType(assignedRequests[i][2])
		elevio.SetButtonLamp(btn_type, btn_floor, true)
		FSM.FSM_onRequestButtonPress(btn_floor, btn_type)
	}
}

func Primary_HandleReceivedRequest(message string) {

	receivedRequests := hra.HRA_StringToRequestMatrix(message)
	unassignedRequests, allRequests := hra.HRA_SortRecievedRequests(receivedRequests)

	utils.HallreqInstrMutex.Lock()
	utils.Utils_AddToRequestMap(allRequests)
	utils.HallreqInstrMutex.Unlock()

	utils.MyElevsMutex.Lock()
	for req := range receivedRequests {

		floor := receivedRequests[req][1]
		btn := elevcons.ButtonType(receivedRequests[req][2])

		if receivedRequests[req][0] == elevcons.NewReq {
			utils.MyElev.Lights[floor][btn] = 1
		} else if receivedRequests[req][0] == elevcons.CompletedReq {
			utils.MyElev.Lights[floor][btn] = 0
		}
	}
	utils.MyElevsMutex.Unlock()

	FSM.FSM_SetAllLights()

	assignedRequests := hra.HRA_HallRequestAssigner(hra.HRAElevstates, unassignedRequests[:])

	utils.HallreqInstrMutex.Lock()

	for IP := range utils.CurrentElevs {
		if IP != utils.PrimaryIP {
			utils.HallreqInstructions[IP] = append(utils.HallreqInstructions[IP], assignedRequests[IP]...)
		}
	}
	utils.HallreqInstrMutex.Unlock()

	executeAssignedRequests(assignedRequests[utils.MyIP])
}

func Primary_HandleButtonsPress(btnEvent elevcons.ButtonEvent) {

	elevio.SetButtonLamp(btnEvent.Button, btnEvent.Floor, true)

	utils.MyElevsMutex.Lock()
	utils.MyElev.Lights[btnEvent.Floor][btnEvent.Button] = 1
	utils.MyElevsMutex.Unlock()

	utils.HallreqInstrMutex.Lock()
	for key := range utils.HallreqInstructions {
		if key != utils.PrimaryIP {
			utils.HallreqInstructions[key] = append(utils.HallreqInstructions[key], [3]int{elevcons.TurnOnLight, btnEvent.Floor, int(btnEvent.Button)})
		}
	}
	utils.HallreqInstrMutex.Unlock()

	newRequest := [][3]int{{elevcons.NewReq, btnEvent.Floor, int(btnEvent.Button)}}
	unassignedRequests, _ := hra.HRA_SortRecievedRequests(newRequest)
	assignedRequests := hra.HRA_HallRequestAssigner(hra.HRAElevstates, unassignedRequests[:])

	utils.HallreqInstrMutex.Lock()
	for key := range utils.HallreqInstructions {
		utils.HallreqInstructions[key] = append(utils.HallreqInstructions[key], assignedRequests[key]...)
	}
	utils.HallreqInstrMutex.Unlock()

	for req := range assignedRequests[utils.MyIP] {
		btn_floor := assignedRequests[utils.MyIP][req][1]
		btn_type := elevcons.ButtonType(assignedRequests[utils.MyIP][req][2])
		FSM.FSM_onRequestButtonPress(btn_floor, btn_type)
	}
}
