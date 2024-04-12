package main

import (
	"Project/clientElevator"
)

func main() {
	
	go clientElevator.ClientMain()
	
	select{}
}
