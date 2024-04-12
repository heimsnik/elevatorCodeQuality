package elevator

import (
	"Heisprosjekt/elevcons"
)

type Elevator struct {
	CurrentFloor int
	Direction    elevcons.MotorDirection
	Behaviour    elevcons.ElevatorBehaviour
	Requests     [elevcons.N_Floors][elevcons.N_Buttons]int
	Status       elevcons.ElevatorMode
	ReturnStatus int
	Lights       [elevcons.N_Floors][elevcons.N_Buttons]int
}

type DirectionBehaviourPair struct {
	Direction elevcons.MotorDirection
	Behaviour elevcons.ElevatorBehaviour
}

func Elevator_GetDirectionBehaviourPair(d elevcons.MotorDirection, b elevcons.ElevatorBehaviour) DirectionBehaviourPair {

	DB := DirectionBehaviourPair{
		Direction: d,
		Behaviour: b,
	}

	return DB
}
