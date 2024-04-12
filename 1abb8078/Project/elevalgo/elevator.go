package elevalgo

import (
	"Project/elevio"
	"fmt"
	"strconv"
	"time"
)

var elevator Elevator // package scope variable 

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	Floor              int
	Dirn               elevio.MotorDirection
	Requests           [elevio.NUM_FLOORS][elevio.NUM_BUTTONS]bool
	Behaviour          ElevatorBehaviour
	doorOpenDuration_s time.Duration
	Obstruction        bool
}

func GetElevator() Elevator {
	return elevator
}

func ebToString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "EB_Idle"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	case EB_Moving:
		return "EB_Moving"
	default:
		return "EB_Unknown"
	}
}

func ElevatorPrint(elev Elevator) {
	fmt.Println("  +--------------------+")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |dirn  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n"+
			"  |obst  = %-12.12s|\n",
		elev.Floor,
		elevio.DirnToString(elev.Dirn),
		ebToString(elev.Behaviour),
		strconv.FormatBool(elev.Obstruction),
	)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := elevio.NUM_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < elevio.NUM_BUTTONS; btn++ {
			if (f == elevio.NUM_FLOORS-1 && btn == int(elevio.BT_HallUp)) ||
				(f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Print("|     ")
			} else {
				if !(elev.Requests[f][btn]) {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

