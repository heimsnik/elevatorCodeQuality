package cab

import (
	"decision"
	"door"
	"elevatorglobals"
	"fmt"
	"location"
	"motor"
	"obstruction"
	"time"
)

func printElevator(cabState elevatorglobals.CabState) {
	fmt.Printf("Cab Floor: %d\n", cabState.Floor)
	fmt.Printf("Direction: %d\n", cabState.Direction)
	fmt.Printf("Behaviour: %d\n", cabState.Behaviour)
	fmt.Printf("--------------------\n")
}

func printAssignedOrders(assignedOrders elevatorglobals.AssignedOrders) {
	hasAnyOrders := false
	for floor, buttons := range assignedOrders {
		for button, assigned := range buttons {
			if assigned {
				fmt.Printf("Floor: %d, Button: %d\n", floor, button)
				hasAnyOrders = true
			}
		}
	}
	if hasAnyOrders {
		fmt.Printf("--------------------\n")
	}
}

func Run(cabState elevatorglobals.CabState,
	assignedOrdersChannel <-chan elevatorglobals.AssignedOrders,
	updateStateChannel chan<- elevatorglobals.CabState,
	orderHandledChannel chan<- elevatorglobals.OrderEvent,
) {

	floorChannel := make(chan int)
	go location.Poll(floorChannel)

	obstructionChannel := make(chan bool)
	go obstruction.Poll(obstructionChannel)

	doorTicker := time.NewTicker(elevatorglobals.DoorOpenDuration)

	assignedOrders := elevatorglobals.AssignedOrders{}
	
	watchdogTimeoutChannel := make(chan bool)
	watchdogDirectionUpdateChannel := make(chan elevatorglobals.Direction)
	watchdogFloorUpdateChannel := make(chan bool)
	watchdogObstructionRemovedChannel := make(chan bool)
	go motor.RunWatchdog(watchdogTimeoutChannel, watchdogDirectionUpdateChannel, watchdogFloorUpdateChannel, watchdogObstructionRemovedChannel)

	for {
		select {
		case <-watchdogTimeoutChannel:
			if !cabState.Obstructed && cabState.Direction != elevatorglobals.Direction_Stop {
				fmt.Println("cab: Watchdog timeout")
				cabState.MotorWorking = false
				updateStateChannel <- cabState
			}

		case floor := <-floorChannel:
			fmt.Println("cab: Floor channel", floor)
			cabState.MotorWorking = true
			watchdogFloorUpdateChannel <- true

			location.SetFloorIndicator(floor)
			cabState.Floor = floor

			if cabState.Behaviour == elevatorglobals.ElevatorBehaviour_Moving && decision.ShouldStop(cabState, assignedOrders) {

				orderHandledList := decision.GetOrdersHandled(cabState, assignedOrders)
				for _, event := range orderHandledList {
					orderHandledChannel <- event
				}

				fmt.Println("cab: Location: floor = ", floor)
				motor.SetDirection(elevatorglobals.Direction_Stop)
				watchdogDirectionUpdateChannel <- cabState.Direction

				door.Toggle(true)
				cabState.Behaviour = elevatorglobals.ElevatorBehaviour_DoorOpen
				fmt.Println("cab: Resetting door timer from location")
				if !cabState.Obstructed {
					doorTicker.Reset(elevatorglobals.DoorOpenDuration)
				}
			}
			updateStateChannel <- cabState

		case isObstructed := <-obstructionChannel:
			if isObstructed {
				doorTicker.Stop()
				fmt.Println("cab: Obstructed")
				cabState.Obstructed = true
			} else {
				doorTicker = time.NewTicker(elevatorglobals.DoorOpenDuration)
				watchdogFloorUpdateChannel <- true
				cabState.Obstructed = false
			}
			updateStateChannel <- cabState

		case newAssignedOrders := <-assignedOrdersChannel:

			assignedOrders = newAssignedOrders

			if cabState.Behaviour == elevatorglobals.ElevatorBehaviour_Idle {
				cabState.Direction, cabState.Behaviour = decision.ChooseDirection(cabState, assignedOrders)

				if cabState.Behaviour == elevatorglobals.ElevatorBehaviour_DoorOpen {
					door.Toggle(true)
					if !cabState.Obstructed {
						doorTicker.Reset(elevatorglobals.DoorOpenDuration)
					}

					orderHandledList := decision.GetOrdersHandled(cabState, assignedOrders)
					for _, event := range orderHandledList {
						orderHandledChannel <- event
					}
				} else if cabState.Behaviour == elevatorglobals.ElevatorBehaviour_Moving {
					motor.SetDirection(cabState.Direction)
					watchdogDirectionUpdateChannel <- cabState.Direction
				}

			}
			updateStateChannel <- cabState

		case <-doorTicker.C:

			if cabState.Behaviour == elevatorglobals.ElevatorBehaviour_DoorOpen {
				fmt.Println("cab: Door timeout")
				newDirection, newBehaviour := decision.ChooseDirection(cabState, assignedOrders)

				if newBehaviour == elevatorglobals.ElevatorBehaviour_DoorOpen {
					fmt.Println("cab: Resetting door timer from door timeout")
					if !cabState.Obstructed {
						doorTicker.Reset(elevatorglobals.DoorOpenDuration)
					}

					cabState.Direction = newDirection
					cabState.Behaviour = newBehaviour
					orderHandledList := decision.GetOrdersHandled(cabState, assignedOrders)
					for _, event := range orderHandledList {
						orderHandledChannel <- event
					}
				} else if newBehaviour == elevatorglobals.ElevatorBehaviour_Moving ||
					newBehaviour == elevatorglobals.ElevatorBehaviour_Idle {
					motor.SetDirection(newDirection)
					watchdogDirectionUpdateChannel <- cabState.Direction

					door.Toggle(false)
					cabState.Direction = newDirection
					cabState.Behaviour = newBehaviour
				}

			}

			updateStateChannel <- cabState
		}
	}
}
