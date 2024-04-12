package backup

import (
	"Heisprosjekt/FSM"
	"Heisprosjekt/driver-go-master/elevio"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/hra"
	"Heisprosjekt/tcp"
	"Heisprosjekt/utils"
	"time"
)

func Backup_HandleReceivedRequest(message string) {

	receivedRequest := hra.HRA_StringToRequestMatrix(message)

	for _, req := range receivedRequest {

		btnFloor := req[1]
		btnType := elevcons.ButtonType(req[2])

		switch req[0] {
		case elevcons.TurnOffLight:
			elevio.SetButtonLamp(btnType, btnFloor, false)

			utils.MyElevsMutex.Lock()
			utils.MyElev.Lights[btnFloor][btnType] = 0
			utils.MyElevsMutex.Unlock()

		case elevcons.TurnOnLight:
			elevio.SetButtonLamp(btnType, btnFloor, true)

			utils.MyElevsMutex.Lock()
			utils.MyElev.Lights[btnFloor][btnType] = 1
			utils.MyElevsMutex.Unlock()

		case elevcons.TakeReq:
			elevio.SetButtonLamp(btnType, btnFloor, true)

			utils.MyElevsMutex.Lock()
			utils.MyElev.Lights[btnFloor][btnType] = 1
			utils.MyElevsMutex.Unlock()

			FSM.FSM_onRequestButtonPress(btnFloor, btnType)
		}
	}
}

func Backup_HallRequestSender() {

	for {

		if utils.MyElev.Status == elevcons.Backup && len(utils.HallreqInstructions[utils.PrimaryIP]) != 0 {
			tcp.TCP_MessageSender(utils.PrimaryIP+elevcons.TcpPort, hra.HRA_RequestMatrixToString(utils.HallreqInstructions[utils.PrimaryIP]))

			utils.HallreqInstrMutex.Lock()
			utils.HallreqInstructions[utils.PrimaryIP] = [][3]int{}
			utils.HallreqInstrMutex.Unlock()
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func Backup_HandleButtonPress(btnEvent elevcons.ButtonEvent) {

	btnFloor := btnEvent.Floor
	btnType := int(btnEvent.Button)
	newReq := [3]int{elevcons.NewReq, btnFloor, btnType}

	utils.HallreqInstrMutex.Lock()
	utils.HallreqInstructions[utils.PrimaryIP] = append(utils.HallreqInstructions[utils.PrimaryIP], newReq)
	utils.HallreqInstrMutex.Unlock()

}
