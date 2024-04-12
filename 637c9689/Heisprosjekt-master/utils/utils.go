package utils

import (
	"Heisprosjekt/elevator"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/network/localip"
	"fmt"
	"sync"
	"time"
)

var (
	MyElev              elevator.Elevator
	MyIP                string
	PrimaryIP           string
	TimerVariable       time.Timer
	MotorCrashTimer     *time.Timer
	CurrentElevs        = map[string]elevator.Elevator{}
	NeedInfoElevs       = map[string]elevator.Elevator{}
	LostHallCalls       = [][3]int{}
	HallreqInstructions = map[string][][3]int{}
)

var (
	MyElevsMutex       = sync.Mutex{}
	PrimatyIPMutex     = sync.Mutex{}
	HallreqInstrMutex  = sync.Mutex{}
	CurrentElevsMutex  = sync.Mutex{}
	NeedInfoElevsMutex = sync.Mutex{}
)

func Utils_GetIP() string {
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	return localIP
}

func Utils_AddToRequestMap(req [][3]int) {
	if MyElev.Status == elevcons.Backup {
		HallreqInstructions[PrimaryIP] = append(HallreqInstructions[PrimaryIP], req...)
		return
	}

	for IP := range HallreqInstructions {
		if IP != PrimaryIP {
			HallreqInstructions[IP] = append(HallreqInstructions[IP], req...)
		}
	}
}

func Utils_IntToBool(i int) bool {
	return (i != 0)
}
