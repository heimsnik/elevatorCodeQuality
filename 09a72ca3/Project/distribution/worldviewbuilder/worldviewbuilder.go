package worldviewbuilder

import (
	"buttons"
	"elevatorglobals"
	"fmt"
	"lamps"
	"network/bcast"
	"network/peers"
	"strings"
	"time"
)

func printWorldView(worldview elevatorglobals.Worldview) {
	fmt.Println("worldviewbuilder: Current worldview")
	fmt.Println("Cabs:")
	for i := 0; i < worldview.ElevatorCount(); i++ {
		fmt.Printf("Elevator %s: ", worldview.CabStates[i].Name)
		fmt.Printf("Floor: %d, ", worldview.CabStates[i].Floor)
		fmt.Printf("Direction: %d, ", worldview.CabStates[i].Direction)
		fmt.Printf("Behaviour: %d, ", worldview.CabStates[i].Behaviour)
		fmt.Printf("Obstructed: %t\n", worldview.CabStates[i].Obstructed)
		fmt.Printf("Motor working: %t\n", worldview.CabStates[i].MotorWorking)
		fmt.Printf("Online: %t\n", worldview.CabStates[i].Online)
	}
	fmt.Println("Hall orders:")
	for i := 0; i < elevatorglobals.FloorCount; i++ {
		fmt.Printf("Floor %d: ", i)
		fmt.Printf("Up: %t, ", worldview.HallOrders[i][0])
		fmt.Printf("Down: %t\n", worldview.HallOrders[i][1])
	}
	fmt.Println("Cab orders:")
	for i := 0; i < worldview.ElevatorCount(); i++ {
		fmt.Printf("Elevator %d: ", i)
		for j := 0; j < elevatorglobals.FloorCount; j++ {
			fmt.Printf("%t, ", worldview.CabOrders[i][j])
		}
		fmt.Println()
	}
	fmt.Println("--------------------")
}

func printPeerUpdate(peerUpdate peers.PeerUpdate) {
	fmt.Printf("worldviewbuilder: Peer update:\n")
	fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
	fmt.Printf("  New:      %q\n", peerUpdate.New)
	fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
	fmt.Println("--------------------")
}

func Run(cabStateChannel <-chan elevatorglobals.CabState,
	orderHandledChannel <-chan elevatorglobals.OrderEvent,
	peerUpdateChannel <-chan peers.PeerUpdate,
	worldviewOutputChannel chan<- elevatorglobals.Worldview) {

	worldviewNetworkTx := make(chan elevatorglobals.WorldviewUpdate)
	worldviewNetworkRx := make(chan elevatorglobals.WorldviewUpdate)
	go bcast.Transmitter(54321, worldviewNetworkTx)
	go bcast.Receiver(54321, worldviewNetworkRx)

	orderHandledTx := make(chan elevatorglobals.OrderEvent)
	orderHandledRx := make(chan elevatorglobals.OrderEvent)
	go bcast.Transmitter(54322, orderHandledTx)
	go bcast.Receiver(54322, orderHandledRx)

	buttonPressedChannel := make(chan elevatorglobals.Order)
	go buttons.Poll(buttonPressedChannel)

	myWorldview := elevatorglobals.Worldview{}
	var onlinePeers []string

	worldviewPublishTicker := time.NewTicker(50 * time.Millisecond)
	worldviewPrintTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case event := <-buttonPressedChannel:
			myWorldview = addOrderToWorldview(myWorldview, event, elevatorglobals.MyElevatorName)

		case event := <-cabStateChannel:
			myWorldview = updateCabState(myWorldview, event)

		case worldviewUpdate := <-worldviewNetworkRx:
			myWorldview = mergeWorldviews(myWorldview, worldviewUpdate.Worldview, worldviewUpdate.OriginName, onlinePeers)

		case OrderEvent := <-orderHandledChannel:
			// Resend orderHandled events while door is open for static redundancy
			resendDuration := elevatorglobals.DoorOpenDuration - 500*time.Millisecond
			resendPeriod := 30 * time.Millisecond
			resendCount := resendDuration / (resendPeriod)

			go func() {
				for range resendCount {
					orderHandledTx <- OrderEvent
					time.Sleep(resendPeriod)
				}
			}()

			myWorldview = handleOrder(myWorldview, OrderEvent)

		case OrderEvent := <-orderHandledRx:
			myWorldview = handleOrder(myWorldview, OrderEvent)

		case <-worldviewPrintTicker.C:
			printWorldView(myWorldview)

		case <-worldviewPublishTicker.C:
			myWorldview = setCabOnlineStatus(myWorldview, onlinePeers)
			lamps.Update(myWorldview)
			worldviewNetworkTx <- elevatorglobals.WorldviewUpdate{Worldview: myWorldview, OriginName: elevatorglobals.MyElevatorName}
			worldviewOutputChannel <- myWorldview

		case peerUpdate := <-peerUpdateChannel:

			allOnlinePeers := peerUpdate.Peers
			validOnlinePeers := []string{}
			for i := range allOnlinePeers {
				if strings.Contains(allOnlinePeers[i], elevatorglobals.Codeword) {
					validOnlinePeers = append(validOnlinePeers, allOnlinePeers[i])
				}
			}
			onlinePeers = validOnlinePeers

			// Our elevator should always be considered online, even though the packet loss script may cause it to not be in the peer list
			myElevatorIsOnline := false
			for _, peer := range onlinePeers {
				if peer == elevatorglobals.MyElevatorName {
					myElevatorIsOnline = true
					break
				}
			}
			if !myElevatorIsOnline {
				onlinePeers = append(onlinePeers, elevatorglobals.MyElevatorName)
			}
			printPeerUpdate(peerUpdate)
		}
	}
}

func addOrderToWorldview(worldview elevatorglobals.Worldview, Order elevatorglobals.Order, OriginName string) elevatorglobals.Worldview {
	if Order.Button == elevatorglobals.ButtonType_Cab {
		OriginIndex := worldview.ElevatorIndex(OriginName)
		if !worldview.CabStates[OriginIndex].Obstructed && worldview.CabStates[OriginIndex].MotorWorking {
			worldview.CabOrders[OriginIndex][Order.Floor] = true
		}
		if Order.Floor == worldview.CabStates[OriginIndex].Floor {
			worldview.CabOrders[OriginIndex][Order.Floor] = true
		}
	} else {
		worldview.HallOrders[Order.Floor][int(Order.Button)] = true
	}
	return worldview
}

func setCabOnlineStatus(worldview elevatorglobals.Worldview, onlinePeers []string) elevatorglobals.Worldview {
	for elevatorIndex := range worldview.ElevatorCount() {
		online := false
		for _, peer := range onlinePeers {
			if peer == worldview.ElevatorNames[elevatorIndex] {
				online = true
				break
			}
		}
		worldview.CabStates[elevatorIndex].Online = online
	}
	return worldview
}

func updateCabState(worldview elevatorglobals.Worldview, cabState elevatorglobals.CabState) elevatorglobals.Worldview {
	if !strings.Contains(cabState.Name, elevatorglobals.Codeword) {
		fmt.Println("worldviewbuilder: elevator name doesn't contain codeword. Ignoring cab state update.")
		return worldview
	}

	elevatorIndex := worldview.ElevatorIndex(cabState.Name)
	if elevatorIndex == -1 {
		worldview.CabStates[worldview.ElevatorCount()] = cabState
		worldview.ElevatorNames[worldview.ElevatorCount()] = cabState.Name
	} else {
		worldview.CabStates[elevatorIndex] = cabState
	}

	return worldview
}

func updateCabOrders(worldview elevatorglobals.Worldview, cabState elevatorglobals.CabState, cabOrders [elevatorglobals.FloorCount]bool) elevatorglobals.Worldview {
	if !strings.Contains(cabState.Name, elevatorglobals.Codeword) {
		fmt.Println("worldviewbuilder: elevator name doesn't contain codeword. Ignoring cab order update.")
		return worldview
	}

	elevatorIndex := worldview.ElevatorIndex(cabState.Name)
	if elevatorIndex != -1 {
		worldview.CabOrders[elevatorIndex] = cabOrders
	} else {
		fmt.Println("worldviewbuilder: elevator name not found in worldview. Ignoring cab order update")
	}

	return worldview
}

func mergeWorldviews(myWorldview elevatorglobals.Worldview, otherWorldview elevatorglobals.Worldview, otherElevatorName string, onlinePeers []string) elevatorglobals.Worldview {
	mergedWorldview := myWorldview

	for elevatorIndex_Other := range otherWorldview.ElevatorCount() {
		if mergedWorldview.ElevatorIndex(otherWorldview.ElevatorNames[elevatorIndex_Other]) == -1 {
			mergedWorldview = updateCabState(mergedWorldview, otherWorldview.CabStates[elevatorIndex_Other])
			mergedWorldview = updateCabOrders(mergedWorldview, otherWorldview.CabStates[elevatorIndex_Other], otherWorldview.CabOrders[elevatorIndex_Other])
		}
	}

	// Other elevator has full authority over its state
	otherElevatorIndex_Other := otherWorldview.ElevatorIndex(otherElevatorName)
	if otherElevatorIndex_Other != -1 {
		mergedWorldview = updateCabState(mergedWorldview, otherWorldview.CabStates[otherElevatorIndex_Other])
	}

	for elevatorIndex_Other := range otherWorldview.ElevatorCount() {
		elevatorIndex_Merged := mergedWorldview.ElevatorIndex(otherWorldview.ElevatorNames[elevatorIndex_Other])
		if elevatorIndex_Merged == -1 {
			mergedWorldview = updateCabState(mergedWorldview, otherWorldview.CabStates[elevatorIndex_Other])
			mergedWorldview = updateCabOrders(mergedWorldview, otherWorldview.CabStates[elevatorIndex_Other], otherWorldview.CabOrders[elevatorIndex_Other])
		} else {
			for floorIndex := range elevatorglobals.FloorCount {
				if otherWorldview.CabOrders[elevatorIndex_Other][floorIndex] {
					mergedWorldview.CabOrders[elevatorIndex_Merged][floorIndex] = true
				}
			}
		}
	}

	for floorIndex := range elevatorglobals.FloorCount {
		if otherWorldview.HallOrders[floorIndex][0] {
			mergedWorldview.HallOrders[floorIndex][0] = true
		}
		if otherWorldview.HallOrders[floorIndex][1] {
			mergedWorldview.HallOrders[floorIndex][1] = true
		}
	}

	return mergedWorldview
}

func handleOrder(worldview elevatorglobals.Worldview, OrderEvent elevatorglobals.OrderEvent) elevatorglobals.Worldview {
	if OrderEvent.Button == elevatorglobals.ButtonType_Cab {
		OriginIndex := worldview.ElevatorIndex(OrderEvent.OriginName)
		if OriginIndex == -1 {
			return worldview
		}
		worldview.CabOrders[OriginIndex][OrderEvent.Floor] = false

	} else {
		worldview.HallOrders[OrderEvent.Floor][OrderEvent.Button] = false
	}

	return worldview
}
