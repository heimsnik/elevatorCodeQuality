package assignerwrapper

import (
	"bytes"
	"elevatorglobals"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func Run(filteredWorldviewChannel <-chan elevatorglobals.Worldview, assignedOrdersChannel chan<- elevatorglobals.AssignedOrders) {
	for worldview := range filteredWorldviewChannel {
		if !isValidWorldview(worldview) {
			myElevatorIndex := worldview.ElevatorIndex(elevatorglobals.MyElevatorName)
			if myElevatorIndex != -1 {
				assignedOrders := elevatorglobals.AssignedOrders{}
				for floorIndex := range worldview.CabOrders[myElevatorIndex] {
					assignedOrders[floorIndex][elevatorglobals.ButtonType_Cab] = worldview.CabOrders[myElevatorIndex][floorIndex]
				}
				assignedOrdersChannel <- assignedOrders
			}

			continue
		}

		programFlags := "--includeCab --input '" + worldviewToJSON(worldview) + "'"
		cmd := exec.Command("bash", "-c", "hall_request_assigner "+programFlags)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Start()
		if err != nil {
			fmt.Println("error starting hall_request_assigner")
			fmt.Println(out.String())
			log.Fatal(err)
		}

		err = cmd.Wait()
		if err != nil {
			fmt.Println("error waiting for hall_request_assigner to finish")
			fmt.Println(out.String())
			log.Fatal(err)
		}

		outputString := out.String()
		allAssignedOrders := make(map[string]elevatorglobals.AssignedOrders)
		json.Unmarshal([]byte(outputString), &allAssignedOrders)
		assignedOrders := allAssignedOrders[elevatorglobals.MyElevatorName]
		assignedOrdersChannel <- assignedOrders

	}
}

type jsonInputState struct {
	Behaviour   string                           `json:"behaviour"` // "idle", "moving", "doorOpen"
	Floor       int                              `json:"floor"`
	Direction   string                           `json:"direction"` // "up", "down", "stop"
	CabRequests [elevatorglobals.FloorCount]bool `json:"cabRequests"`
}

type jsonInput struct {
	HallRequests [elevatorglobals.FloorCount][2]bool `json:"hallRequests"`
	States       map[string]jsonInputState           `json:"states"`
}

var jsonBehaviourStringMap = map[elevatorglobals.ElevatorBehaviour]string{
	elevatorglobals.ElevatorBehaviour_Idle:     "idle",
	elevatorglobals.ElevatorBehaviour_DoorOpen: "doorOpen",
	elevatorglobals.ElevatorBehaviour_Moving:   "moving",
}

var JsonDirectionStringMap = map[elevatorglobals.Direction]string{
	elevatorglobals.Direction_Up:   "up",
	elevatorglobals.Direction_Down: "down",
	elevatorglobals.Direction_Stop: "stop",
}

func worldviewToJSON(worldview elevatorglobals.Worldview) string {
	jsonInput := jsonInput{
		HallRequests: worldview.HallOrders,
		States:       map[string]jsonInputState{},
	}
	for elevatorIndex := range worldview.ElevatorCount() {
		if !worldview.CabStates[elevatorIndex].MotorWorking ||
			worldview.CabStates[elevatorIndex].Obstructed ||
			!worldview.CabStates[elevatorIndex].Online {
			continue
		}
		cabState := worldview.CabStates[elevatorIndex]
		jsonInput.States[worldview.ElevatorNames[elevatorIndex]] = jsonInputState{
			Behaviour:   jsonBehaviourStringMap[cabState.Behaviour],
			Floor:       cabState.Floor,
			Direction:   JsonDirectionStringMap[cabState.Direction],
			CabRequests: worldview.CabOrders[elevatorIndex],
		}
	}

	jsonBytes, err := json.Marshal(jsonInput)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	jsonString := string(jsonBytes)

	return jsonString
}

func isValidWorldview(worldview elevatorglobals.Worldview) bool {
	workingElevators := 0
	for elevatorIndex := range worldview.ElevatorCount() {
		if worldview.CabStates[elevatorIndex].MotorWorking && !worldview.CabStates[elevatorIndex].Obstructed && worldview.CabStates[elevatorIndex].Online {
			workingElevators++
		}
	}
	if workingElevators == 0 {
		return false
	}
	myElevatorIndex := worldview.ElevatorIndex(elevatorglobals.MyElevatorName)
	if myElevatorIndex == -1 {
		return false
	}
	if !(worldview.CabStates[myElevatorIndex].MotorWorking && !worldview.CabStates[myElevatorIndex].Obstructed && worldview.CabStates[myElevatorIndex].Online) {
		return false
	}
	return true
}
