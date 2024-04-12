package elevatorglobals

import "time"

const FloorCount = 4
const ButtonCount = 3
const DoorOpenDuration = 3 * time.Second

var MyElevatorName string

const Codeword = "dune"

const maxElevators = 10

type ButtonType int

const (
	ButtonType_HallUp   ButtonType = 0
	ButtonType_HallDown            = 1
	ButtonType_Cab                 = 2
)

var ButtonStringMap = map[ButtonType]string{
	ButtonType_HallUp:   "ButtonType_HallUp",
	ButtonType_HallDown: "ButtonType_HallDown",
	ButtonType_Cab:      "ButtonType_Cab",
}

type Order struct {
	Floor  int
	Button ButtonType
}

type OrderEvent struct {
	OriginName string
	Floor    int
	Button   ButtonType
}

type ElevatorBehaviour int

const (
	ElevatorBehaviour_Idle     ElevatorBehaviour = 0
	ElevatorBehaviour_DoorOpen ElevatorBehaviour = 1
	ElevatorBehaviour_Moving   ElevatorBehaviour = 2
)

type Direction int

const (
	Direction_Up   Direction = 1
	Direction_Down Direction = -1
	Direction_Stop Direction = 0
)

type CabState struct {
	Name           string
	Behaviour    ElevatorBehaviour
	Floor        int
	Direction    Direction
	MotorWorking bool
	Obstructed   bool
	Online       bool
}

type Worldview struct {
	ElevatorNames  [maxElevators]string
	CabStates    [maxElevators]CabState
	CabOrders  [maxElevators][FloorCount]bool
	HallOrders [FloorCount][2]bool
}

func (worldview Worldview) ElevatorCount() int {
	elevatorCount := 0
	for _, name := range worldview.ElevatorNames {
		if name != "" {
			elevatorCount++
		}
	}
	return elevatorCount
}

func (worldview Worldview) ElevatorIndex(name string) int {
	for i, elevatorName := range worldview.ElevatorNames {
		if elevatorName == name {
			return i
		}
	}
	return -1
}

type WorldviewUpdate struct {
	Worldview Worldview
	OriginName  string
}

type AssignedOrders [FloorCount][ButtonCount]bool

func (assignedOrders AssignedOrders) IsEmpty() bool {
	for floor := range FloorCount {
		for button := range ButtonCount {
			if assignedOrders[floor][button] {
				return false
			}
		}
	}
	return true
}
