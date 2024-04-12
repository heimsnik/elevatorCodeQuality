package runelevator

import (
	"Heisprosjekt/FSM"
	"Heisprosjekt/backup"
	"Heisprosjekt/driver-go-master/elevio"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/primary"
	"Heisprosjekt/utils"
	"time"
)

func RunElevator(tcpReceiver chan string, drv_buttons chan elevcons.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool, peerEnableCh chan bool) {

	for {
		select {
		case a := <-tcpReceiver:
			if utils.MyElev.Status == elevcons.Primary {
				primary.Primary_HandleReceivedRequest(a)
			} else if utils.MyElev.Status == elevcons.Backup {
				backup.Backup_HandleReceivedRequest(a)
			}

		case a := <-drv_buttons:
			if a.Button == elevcons.BT_Cab || utils.MyElev.Status == elevcons.SingleElevator {
				utils.MyElev.Lights[a.Floor][a.Button] = 1
				FSM.FSM_onRequestButtonPress(a.Floor, a.Button)
			} else {
				if utils.MyElev.Status == elevcons.Primary {
					primary.Primary_HandleButtonsPress(a)
				} else if utils.MyElev.Status == elevcons.Backup {
					backup.Backup_HandleButtonPress(a)
				}
			}

		case a := <-drv_floors:
			if a != -1 {
				FSM.FSM_OnFloorArrival(a)
			}
		case a := <-drv_obstr:
			for a {
				elevio.SetMotorDirection(elevcons.MD_Stop)
				if elevio.GetFloor() != -1 {
					elevio.SetDoorOpenLamp(true)
				}
				a = <-drv_obstr
			}
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(utils.MyElev.Direction)
		case a := <-drv_stop:
			if a {
				elevio.SetStopLamp(true)
				elevio.SetMotorDirection(elevcons.MD_Stop)
				peerEnableCh <- false
				utils.MyElev.ReturnStatus = 2
				utils.MotorCrashTimer.Stop()
			} else {
				elevio.SetStopLamp(false)
				elevio.SetMotorDirection(utils.MyElev.Direction)
				peerEnableCh <- true
				if utils.MyElev.Direction != elevcons.MD_Stop {
					utils.MotorCrashTimer = time.NewTimer(elevio.MotorCrashTime)
				}
			}
		case <-utils.TimerVariable.C:
			FSM.FSM_OnDoorTimeout()

		case <-utils.MotorCrashTimer.C:
			peerEnableCh <- false
			<-drv_floors
			peerEnableCh <- true
		}
	}
}
