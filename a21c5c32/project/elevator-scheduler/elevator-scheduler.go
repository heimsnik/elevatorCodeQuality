// package elevatorscheduler
package elevsched

import (
	"encoding/json"
	"os/exec"
	"fmt"

	"elevator/structs"
)

// Create the argument in the correct format for the cost function
func assembleArgument(systemData structs.SystemData) MessageStruct {

	// Create empty struct to store data
	new_argument := MessageStruct{}

	// Add button calls to 2x4 array HallRequests
	up := systemData.UP_BUTTON_ARRAY
	down := systemData.DOWN_BUTTON_ARRAY
	requests := [4][2]bool{}

	for i := 0; i < len(up); i++ {
		requests[i][0] = up[i]
		requests[i][1] = down[i]
	}

	new_argument.HallRequests = requests

	// Assemble states
	direction_strings := [3]string{"up", "down", "stop"}
	elevator_strings := [structs.N_ELEVATORS]string{"one", "two", "three"}
	new_states := make(map[string]singleElevatorState)

	for i := 0; i < len(*systemData.ELEVATOR_DATA); i++ {

		// Don't take elevator into account if not alive, stopped or obstructed
		is_stopped := systemData.ELEVATOR_DATA[i].INTERNAL_STATE == structs.STOPPED
		is_obstructed := systemData.ELEVATOR_DATA[i].INTERNAL_STATE == structs.OBSTRUCTED
		if !systemData.ELEVATOR_DATA[i].ALIVE || is_stopped || is_obstructed {
			continue
		}

		new_state := singleElevatorState{}

		state := (*systemData.ELEVATOR_DATA)[i]

		new_state.Behaviour = state_to_behaviour(state)
		new_state.Floor = state.CURRENT_FLOOR
		new_state.Direction = direction_strings[state.DIRECTION]
		new_state.CabRequests = (*systemData.ELEVATOR_DATA)[i].INTERNAL_BUTTON_ARRAY

		// Set the values for the corresponding elevator
		new_states[elevator_strings[i]] = new_state

	}

	// Set the new states
	new_argument.States = new_states

	return new_argument
}

// Translate the elevators state to the corresponding string value
func state_to_behaviour(state structs.ElevatorData) string {
	if state.INTERNAL_STATE == structs.IDLE {
		return "idle"
	}
	if state.INTERNAL_STATE == structs.MOVING {
		return "moving"
	}
	if state.INTERNAL_STATE == structs.DOOR_OPEN {
		return "doorOpen"
	}

	// Any other state is unexpected behaviour
	fmt.Errorf("Unknown internal state reached")
	return ""
}

// Structure containing the data for each elevator
type singleElevatorState struct {
	Behaviour   string  `json:"behaviour"`
	Floor       int     `json:"floor"`
	Direction   string  `json:"direction"`
	CabRequests [4]bool `json:"cabRequests"`
}

// Structure for the full message
type MessageStruct struct {
	HallRequests [4][2]bool                     `json:"hallRequests"`
	States       map[string]singleElevatorState `json:"states"`
}

// Return the movements of the elevator
func CalculateElevatorMovement(systemData structs.SystemData) *(map[string][structs.N_FLOORS][2]bool) {
	command := "./hall_request_assigner"

	// Create json string from the system data
	new_struct := assembleArgument(systemData)
	new_json, err := json.MarshalIndent(new_struct, "", "\t")

	if err != nil {
		fmt.Printf("%s", err)
		return new(map[string][structs.N_FLOORS][2]bool)
	}

	// Run the cost function to get new orders
	cmd := exec.Command(command, "-i", string(new_json))
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return new(map[string][structs.N_FLOORS][2]bool)
	}

	// Decode the new orders
	output := new(map[string][structs.N_FLOORS][2]bool)
	// ERR:
	json.Unmarshal([]byte(stdout), &output)

	return output
}