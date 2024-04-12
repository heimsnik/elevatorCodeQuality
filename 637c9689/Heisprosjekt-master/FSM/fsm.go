package FSM

import (
	"Heisprosjekt/driver-go-master/elevio"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/requests"
	"Heisprosjekt/utils"
	"time"
)

func FSM_Initialize() {

	utils.MyIP = utils.Utils_GetIP()
	utils.MyElev.ReturnStatus = elevcons.NeedAllInfo
	utils.MyElev.CurrentFloor = 0
	utils.MyElev.Behaviour = 0
	utils.MyElev.Direction = 0

	turnOffAllLights()
	utils.MotorCrashTimer = time.NewTimer(elevio.MotorCrashTime)

	if elevio.GetFloor() == -1 {
		onInitBetweenFloors()
	}
}

func onInitBetweenFloors() {

	elevio.SetMotorDirection(elevcons.MD_Down)
	utils.MyElev.Direction = elevcons.MD_Down
	utils.MyElev.Behaviour = elevcons.Moving

	for {
		if elevio.GetFloor() != -1 {
			elevio.SetMotorDirection(elevcons.MD_Stop)
			utils.MyElev.Direction = elevcons.MD_Stop
			utils.MyElev.Behaviour = elevcons.Idle
			break
		}
	}
}

func FSM_onRequestButtonPress(btn_floor int, btn_type elevcons.ButtonType) {

	utils.MyElevsMutex.Lock()
	switch utils.MyElev.Behaviour {
	case elevcons.Door_open:
		if requests.Requests_ShouldClearImmediatly(btn_floor, btn_type) {
			utils.TimerVariable = *time.NewTimer(elevio.DoorOpenTime)
		} else {
			//a := utils.MyElev.Requests
			utils.MyElev.Requests[btn_floor][2] = 1
		}

	case elevcons.Moving:
		utils.MyElev.Requests[btn_floor][int(btn_type)] = 1

	case elevcons.Idle:
		utils.MyElev.Requests[btn_floor][int(btn_type)] = 1
		DB := requests.Requests_ChooseDirection()
		utils.MyElev.Direction = DB.Direction
		utils.MyElev.Behaviour = DB.Behaviour
		switch DB.Behaviour {
		case elevcons.Door_open:
			elevio.SetDoorOpenLamp(true)
			utils.TimerVariable = *time.NewTimer(elevio.DoorOpenTime)
			requests.Requests_ClearAtCurrentFloor()

		case elevcons.Moving:
			elevio.SetMotorDirection(utils.MyElev.Direction)
			if utils.MyElev.Direction != elevcons.MD_Stop {
				utils.MotorCrashTimer = time.NewTimer(elevio.MotorCrashTime)
			}

		case elevcons.Idle:
			break
		}
	}
	FSM_SetAllLights()
	utils.MyElevsMutex.Unlock()
}

func FSM_SetAllLights() {

	for floor := 0; floor < elevcons.N_Floors; floor++ {
		for btn := 0; btn < elevcons.N_Buttons; btn++ {
			elevio.SetButtonLamp(elevcons.ButtonType(btn), floor, utils.MyElev.Lights[floor][btn] != 0)
		}
	}
}

func turnOffAllLights() {

	for floor := 0; floor < elevcons.N_Floors; floor++ {
		for btn := 0; btn < elevcons.N_Buttons; btn++ {
			utils.MyElev.Lights[floor][btn] = 0
		}
	}
	FSM_SetAllLights()
}

func FSM_OnFloorArrival(newFloor int) {

	utils.HallreqInstrMutex.Lock()
	utils.MyElev.CurrentFloor = newFloor
	elevio.SetFloorIndicator(newFloor)
	utils.MotorCrashTimer.Stop()

	switch utils.MyElev.Behaviour {
	case elevcons.Moving:
		if requests.Requests_ShouldStop() {

			elevio.SetMotorDirection(elevcons.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			utils.TimerVariable = *time.NewTimer(elevio.DoorOpenTime)
			requests.Requests_ClearAtCurrentFloor()
			FSM_SetAllLights()
			utils.MyElev.Behaviour = elevcons.Door_open
			break
		}
		utils.MotorCrashTimer = time.NewTimer(elevio.MotorCrashTime)

	default:
		break
	}
	utils.HallreqInstrMutex.Unlock()
}

func FSM_OnDoorTimeout() {

	utils.HallreqInstrMutex.Lock()
	switch utils.MyElev.Behaviour {
	case elevcons.Door_open:
		DB := requests.Requests_ChooseDirection()
		utils.MyElev.Direction = DB.Direction
		utils.MyElev.Behaviour = DB.Behaviour

		switch utils.MyElev.Behaviour {
		case elevcons.Door_open:
			utils.TimerVariable = *time.NewTimer(elevio.DoorOpenTime)
			requests.Requests_ClearAtCurrentFloor()
			FSM_SetAllLights()

		case elevcons.Moving, elevcons.Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(utils.MyElev.Direction)
			if utils.MyElev.Direction != elevcons.MD_Stop {
				utils.MotorCrashTimer = time.NewTimer(elevio.MotorCrashTime)
			}
		}

	default:
		break
	}
	utils.HallreqInstrMutex.Unlock()
}
