package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"

	master "elevator/master-slave"
	singleelev "elevator/single-elevator"

	"elevator/structs"
	"flag"
	"strconv"
)

func main() {
	fmt.Printf("Hello\n")

	//Gets elevator id from terminal
	var id string
	flag.StringVar(&id, "i", "", "id of this peer")

	// Gets port from terminal
	var port string
	flag.StringVar(&port, "p", "", "port of this peer")

	fmt.Printf("flag: %s\n", id)
	fmt.Printf("port: %s\n", port)
	//Must be after flags, but before the input from flags are used
	flag.Parse()

	received_id, _ := strconv.Atoi(id)

	//Specifies port so that several simulators can be run on same computer
	elevio.Init("localhost:15657", structs.N_FLOORS)

	singleelev.ResetElevator()

	// Initialize the channels for receiving data from the elevio interface
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	// Create master-slave
	master_slave := master.MakeMasterSlave(received_id, ":"+port)

	

	// Create elevator and with corresponding master-slave
	elevator := singleelev.MakeElevator(received_id, master_slave)

	// Start reading elevator channels and the main loop of the elevator
	go elevator.ReadChannels(drv_buttons, drv_floors, drv_obstr, drv_stop)
	go elevator.ElevatorLoop()

	// Start master-slave main loop
	go master_slave.StartMasterSlave()

	// Prevent the program from terminating
	for {
		time.Sleep(time.Minute)
	}

}
