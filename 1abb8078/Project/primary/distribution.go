package primary

import (
	"Project/elevalgo"
	"Project/elevio"
	"Project/network/tcpnetwork"
)

// to weigh the cost of actions in distribution algorithm
const travel_cost int = 4
const door_open_cost int = 3

func elevatorToHandle(btnEvent elevio.ButtonEvent, 
	availableElevators []int) int {

	floor := btnEvent.Floor
	btntyp := btnEvent.Button

	elevatorAndTime := [][]int{}

	// fill in time to handle for each elevator
	for i := 0; i < len(availableElevators); i++ {
		var elev elevalgo.Elevator = elevators[availableElevators[i]]
		var time int = timeToHandle(elev, floor, btntyp)
		elevatorAndTime = append(elevatorAndTime, []int{availableElevators[i], time})
	}

	// find min time
	var minTime int = elevatorAndTime[0][1]
	var minId int = elevatorAndTime[0][0]
	for i := 1; i < len(elevatorAndTime); i++ {
		if elevatorAndTime[i][1] < minTime {
			minTime = elevatorAndTime[i][1]
			minId = elevatorAndTime[i][0]
		}
	}

	return minId
}

func availableElevators(elevator_socket *tcpnetwork.PrimaryToElevatorTCPServer) []int {
	connectedElevators := elevator_socket.GetActiveConnections()
	availableElevators := make([]int, 0)
	for i := 0; i < len(connectedElevators); i++ {
		id := connectedElevators[i]
		if !elevators[id].Obstruction{
			availableElevators = append(availableElevators, id)
		}
	}
	return availableElevators
}

func timeToHandle(elev elevalgo.Elevator, floor int, btntyp elevio.ButtonType) int {
	duration := 0

	elev.Requests[floor][btntyp] = true
	switch elev.Behaviour {
	case elevalgo.EB_Idle:
		behaviorPair := elevalgo.Req_chooseDirection(elev)
		elev.Dirn = behaviorPair.Dirn
		if elev.Dirn == elevio.MD_Stop {
			return duration
		}
	case elevalgo.EB_Moving:
		duration += travel_cost / 2
		elev.Floor += int(elev.Dirn)
	case elevalgo.EB_DoorOpen:
		duration -= door_open_cost / 2
	}

	for {
		if elevalgo.Req_shouldStop(elev) {
			elev = elevalgo.Req_clearAtCurrentFloor(elev)
			duration += door_open_cost
			behaviorPair := elevalgo.Req_chooseDirection(elev)
			elev.Dirn = behaviorPair.Dirn
			if elev.Dirn == elevio.MD_Stop {
				return duration
			}
		}
		elev.Floor += int(elev.Dirn)
		duration += travel_cost
	}
}
