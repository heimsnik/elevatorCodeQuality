package main


import (
	"Driver-go/masterbackup"
	"Driver-go/state"
)

func main() {
  
	// Run one elevator
    state.Run_single_elev(4145)

	masterbackup.Run_master_backup()
}




    


