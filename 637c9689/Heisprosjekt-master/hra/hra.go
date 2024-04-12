package hra

import (
	"Heisprosjekt/elevator"
	"Heisprosjekt/elevcons"
	"Heisprosjekt/utils"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var HRAElevstates = map[string]HRAElevState{}

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func HRA_ElevatorToHRAElevState(e elevator.Elevator) HRAElevState {

	floor := e.CurrentFloor
	behaviour := ""
	direction := ""
	cabRequests := []bool{}

	for floor := 0; floor < elevcons.N_Floors; floor++ {
		cabRequests = append(cabRequests, utils.Utils_IntToBool(e.Requests[floor][2]))
	}

	switch e.Behaviour {
	case elevcons.Door_open:
		behaviour = "doorOpen"
	case elevcons.Idle:
		behaviour = "idle"
	case elevcons.Moving:
		behaviour = "moving"
	default:
		panic(e.Behaviour)
	}

	switch e.Direction {
	case elevcons.MD_Down:
		direction = "down"
	case elevcons.MD_Up:
		direction = "up"
	case elevcons.MD_Stop:
		direction = "stop"
	default:
		panic(e.Direction)
	}

	return HRAElevState{Behavior: behaviour, Floor: floor, Direction: direction, CabRequests: cabRequests}
}

func HRA_UpdateHRAElevstates() {

	for IP, elev := range utils.CurrentElevs {
		HRAElevstates[IP] = HRA_ElevatorToHRAElevState(elev)
	}
	time.Sleep(250 * time.Millisecond)

}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func HRA_HallRequestAssigner(HRAElevstates map[string]HRAElevState, unassignedRequests [][2]bool) map[string][][3]int {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := HRAInput{
		HallRequests: unassignedRequests,
		States:       HRAElevstates,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	ret, err := exec.Command("./hra/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}

	assignedRequests := make(map[string][][3]int)

	for IP, reqList := range *output {
		for floor, btnVec := range reqList {
			for btn := range btnVec {
				if reqList[floor][btn] {
					assignedRequests[IP] = append(assignedRequests[IP], [3]int{elevcons.TakeReq, floor, btn})
				}
			}
		}
	}

	return assignedRequests
}

func HRA_SortRecievedRequests(receivedRequest [][3]int) ([elevcons.N_Floors][2]bool, [][3]int) {

	unassignedRequests := [elevcons.N_Floors][2]bool{{false, false}}
	allRequests := [][3]int{}

	for floor := 0; floor < len(receivedRequest); floor++ {

		if receivedRequest[floor][0] == elevcons.NewReq {
			unassignedRequests[receivedRequest[floor][1]][receivedRequest[floor][2]] = true
			allRequests = append(allRequests, [3]int{elevcons.TurnOnLight, receivedRequest[floor][1], receivedRequest[floor][2]})

		} else if receivedRequest[floor][0] == elevcons.CompletedReq {

			allRequests = append(allRequests, [3]int{elevcons.TurnOffLight, receivedRequest[floor][1], receivedRequest[floor][2]})
		}
	}

	return unassignedRequests, allRequests
}

func HRA_StringToRequestMatrix(hrString string) [][3]int {

	hrArray := [][3]int{}

	requests := strings.Split(hrString, " ")

	for i := 0; i < len(requests); i++ {
		typeFloorBtn := strings.Split(requests[i], ",")
		reqType, _ := strconv.Atoi(typeFloorBtn[0])
		floor, _ := strconv.Atoi(typeFloorBtn[1])
		btn, _ := strconv.Atoi(typeFloorBtn[2])
		hrArray = append(hrArray, [3]int{reqType, floor, btn})
	}

	return hrArray
}

func HRA_RequestMatrixToString(hrArray [][3]int) string {

	hrString := ""

	for f := 0; f < len(hrArray); f++ {
		if f == (len(hrArray) - 1) {
			hrString += fmt.Sprint(hrArray[f][0], ",", hrArray[f][1], ",", hrArray[f][2])
		} else {
			hrString += fmt.Sprint(hrArray[f][0], ",", hrArray[f][1], ",", hrArray[f][2], " ")
		}
	}
	return hrString
}
