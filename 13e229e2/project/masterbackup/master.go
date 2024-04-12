package masterbackup

import (
	"Driver-go/config"
	"Driver-go/datatypes"
	"Driver-go/elevio"
	"Driver-go/network"
	"Driver-go/utilities"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"time"
)

func Run_master_backup() {
	state := "backup"

	var backupOverview []datatypes.Overview
	var backupQueue []datatypes.Queue

	srv_newConn := make(chan net.Conn)
	srv_lossConn := make(chan net.Addr)
	srv_lossMaster := make(chan bool)
	msg_new := make(chan datatypes.ServerMessage, 1024)	//Buffer for more msg
	msg_backup := make(chan datatypes.ServerMessage)

	reconnectTimer := time.NewTimer(config.ReconnectInterval)
	reconnectTimer.Stop()
	
	for {
		switch state {
		case "master":
			fmt.Print(" --- Starting Master --- \n")
			masterQueue := backupQueue
			clientOverview := backupOverview

			go network.Start_server(srv_lossConn, srv_newConn, msg_new)
			go network.UDP_broadcast()
			Spawn_backup()
			reconnectTimer.Reset(config.ReconnectInterval)

			for {
				select {
				case new := <- srv_newConn:
					clientOverview = append(clientOverview, datatypes.Overview{ClientType: "unkown", Addr: new.RemoteAddr(), Conn: new, ConnStatus: true})
					Sync_to_backup(clientOverview, masterQueue)
				case loss := <- srv_lossConn:
					for _, client := range clientOverview {
						//If backup disconnect -> remove from overview and spawn new backup
						if client.ClientType == "backup" && client.Addr == loss {
							clientOverview = Overview_remove_client(clientOverview, loss)
							Spawn_backup()
							break
						//If elevator disconnects -> redistribut order and remove from client overview
						} else if client.ClientType == "elevator" && client.Addr == loss {
							client.ConnStatus = false
							for _, queueTask := range masterQueue {
								if queueTask.SendTo == loss {
									clientOverview = Overview_remove_task(clientOverview, elevio.ButtonEvent{Floor: queueTask.Floor, Button: queueTask.Button}, loss)
									taskSendTo := Choose_best_elevator(clientOverview, queueTask)
									masterQueue = Queue_update_sendTo(masterQueue, queueTask, taskSendTo)
									clientOverview = Overview_add_task(clientOverview, elevio.ButtonEvent{Floor: queueTask.Floor, Button: queueTask.Button}, taskSendTo)
									Sync_to_backup(clientOverview, masterQueue)
								}
							}
							clientOverview = Overview_remove_client(clientOverview, loss)
							break
						}
					}
					Sync_to_backup(clientOverview, masterQueue)
				case newMsg := <- msg_new:
					switch newMsg.Tag {
					case "info":
						for index, client := range clientOverview {
							if client.Addr == newMsg.Addr {
								clientOverview[index].ClientType = newMsg.ClientInfo
								fmt.Println("info: ", clientOverview)
							}
						}
						Sync_to_backup(clientOverview, masterQueue)
					case "updateFloor":
						for index, client := range clientOverview {
							if client.Addr == newMsg.Addr {
								clientOverview[index].CurrentFloor = newMsg.Floor
							}
						}
						Sync_to_backup(clientOverview, masterQueue)
					case "newRequest":
						masterQueue = append(masterQueue, datatypes.Queue{Floor: newMsg.Task.Floor, Button: newMsg.Task.Button, ReceivedFrom: newMsg.Addr})
						Sync_to_backup(clientOverview, masterQueue)
						taskSendTo := Choose_best_elevator(clientOverview, datatypes.Queue{Floor: newMsg.Task.Floor, Button: newMsg.Task.Button, ReceivedFrom: newMsg.Addr})
						masterQueue = Queue_update_sendTo(masterQueue, datatypes.Queue{Floor: newMsg.Task.Floor, Button: newMsg.Task.Button}, taskSendTo)
						clientOverview = Overview_add_task(clientOverview, elevio.ButtonEvent{Floor: newMsg.Task.Floor, Button: newMsg.Task.Button}, taskSendTo)
						Sync_to_backup(clientOverview, masterQueue)
					case "done":
						masterQueue = Queue_remove_task(masterQueue, newMsg.Task)
						clientOverview = Overview_remove_task(clientOverview, elevio.ButtonEvent{Floor: newMsg.Task.Floor, Button: newMsg.Task.Button}, newMsg.Addr)
						Sync_to_backup(clientOverview, masterQueue)
				
					default:
						fmt.Print("Uknown message type. \n")
					}
				case <- reconnectTimer.C:
					for _, task := range masterQueue {
						for _, client := range clientOverview {
							if task.SendTo == client.Addr && !client.ConnStatus {
								srv_lossConn <- client.Addr
							}
						}
					}
				}
			}
		case "backup":
			fmt.Print(" --- Starting Backup --- \n")
			//Search for server and connect to it
			servExist, servAddr := network.Find_server_ip()
			if servExist {
				conn := network.Connect_to_server(servAddr, "backup")
				go network.Listen_to_server(msg_backup, srv_lossMaster, conn)
			} else {
				state = "master"
				break
			}

			for {
				select {
				case <- srv_lossMaster:
					state = "master"
					break
				case newBackup := <- msg_backup:
					backupOverview = newBackup.Overview
					backupQueue = newBackup.Queue
				default:
					if state == "master" {
						break
					}
				}
			}
		}
	}
}

///////////// Backup

func Spawn_backup() {
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go").Run()
	} else if runtime.GOOS == "linux" {
		exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
	} else {
		fmt.Print("There is problem with os.")
	}
}

func Sync_to_backup(ClientOverview []datatypes.Overview, masterQueue []datatypes.Queue) {
	for _, client := range ClientOverview {
		if client.ClientType == "backup" {
			msg := datatypes.ServerMessage{Tag: "sync", Queue: masterQueue, Overview: ClientOverview}
			network.Send_message(msg, client.Conn)
		}
	}
}

///////////// Queue

func Queue_remove_task(masterQueue []datatypes.Queue, order elevio.ButtonEvent) []datatypes.Queue {
	var temp []datatypes.Queue
	for _, task := range masterQueue {
		if task.Button != order.Button && task.Floor != order.Floor {
			temp = append(temp, task)
		}
	}
	return temp
}

func Queue_update_sendTo(masterQueue []datatypes.Queue, order datatypes.Queue, addr net.Addr) []datatypes.Queue {
	for index, task := range masterQueue {
		if task.Button == order.Button && task.Floor == order.Floor {
			masterQueue[index].SendTo = addr
		}
	}
	return masterQueue
}

///////////// Overview

func Overview_remove_client(clientOverview []datatypes.Overview, addr net.Addr) []datatypes.Overview {
	var temp []datatypes.Overview
	for _, client := range clientOverview {
		if client.Addr != addr {
			temp = append(temp, client)
		}
	}
	return temp
}

func Overview_add_task(ClientOverview []datatypes.Overview, order elevio.ButtonEvent, addr net.Addr) []datatypes.Overview {
	for _, client := range ClientOverview {
		if client.Addr == addr {
			client.Task = append(client.Task, order)
		}
	}
	return ClientOverview
}

func Overview_remove_task(ClientOverview []datatypes.Overview, order elevio.ButtonEvent, addr net.Addr) []datatypes.Overview {
	for index, client := range ClientOverview {
		if client.Addr == addr {
			var temp []elevio.ButtonEvent
			for _, task := range client.Task {
				if task.Button != order.Button && task.Floor != order.Floor {
					temp = append(temp, task)
				}
			}
			ClientOverview[index].Task = temp
		}
	}
	return ClientOverview
}

///////////// Distribute task

func Choose_best_elevator(ClientOverview []datatypes.Overview, order datatypes.Queue) net.Addr {
	cost := make([]int, len(ClientOverview), config.NumElevators * 2)
	//If cab order -> send to the same elevator if connected
	if order.Button == elevio.BT_Cab {
		for _, client := range ClientOverview {
			if client.Addr == order.ReceivedFrom && client.ConnStatus && client.ClientType == "elevator" {
				network.Send_message(datatypes.ServerMessage{Tag: "task", Task: elevio.ButtonEvent{Floor: order.Floor, Button: order.Button}}, client.Conn)
				return client.Addr
			} 
		}
	//Else calculate cost for each elevator, based on distans each elevator need to run, and send that task to cheapes elevator
	} else {
		for index, client := range ClientOverview {
			if client.ConnStatus && client.ClientType == "elevator" {
				position := client.CurrentFloor
				for _, task := range client.Task {
					if order.Floor > client.CurrentFloor && order.Button == elevio.BT_HallUp {
						if order.Floor > task.Floor {
							cost[index] += utilities.Abs_diff(position, task.Floor)
						} else {
							cost[index] += utilities.Abs_diff(position, order.Floor)
						} 
					} else if order.Floor < task.Floor && order.Button == elevio.BT_HallDown {
						if order.Floor < task.Floor {
							cost[index] += utilities.Abs_diff(position, task.Floor)
						} else {
							cost[index] += utilities.Abs_diff(position, order.Floor)
						}
					} else {
						cost[index] += utilities.Abs_diff(position, task.Floor)
					}
				position = task.Floor
				}
			}
		}
	}
	indexMinCost := utilities.Find_Index_of_min(cost)
	network.Send_message(datatypes.ServerMessage{Tag: "task", Task: elevio.ButtonEvent{Floor: order.Floor, Button: order.Button}}, ClientOverview[indexMinCost].Conn)
	//Need to start timer for that task
	return ClientOverview[indexMinCost].Addr
}
// unused
// func Master_backup_loop() {
// 	state := "backup"

// 	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:6000")
// 	if err != nil {
// 		fmt.Println("Something went wrong: ", err)
// 		return
// 	}

// 	//Need to broadcast own ip on special port

// 	//Make queue (floor: int, btn type: elevio.ButtonType, receivedFrom: int, sendTo: int) rec and send is id of elevator
// 	//MasterQueue []state.ElevMessage

// 	//Make elevator overview (id: int, busy: bool, current floor: int, array of task of type ButtonEvent, connection status: bool)
// 	//ElevOverview []

// 	//Master - backup pingpong communication

// 	for {
// 		switch state {
// 		case "master":
// 			fmt.Print("Master loop")
			
// 			//Spawn backup
// 			if runtime.GOOS == "windows" {
// 				exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go").Run()
// 			} else if runtime.GOOS == "linux" {
// 				exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
// 			} else {
// 				fmt.Print("There is problem with os.")
// 			}

// 			//Brodcast msg on udpMAsterPort
// 			//Listen for msg from backup

// 			//Based on queue updae overview

// 			//Start server and listen to clients
// 			server := network.NewServer(":3000")
// 			go func() {
// 				for msg := range server.msgch {
// 					fmt.Printf("Received message from connection (%s):%s \n ", msg.from, string(msg.payload))
// 					//HandleMsg function that will do:
// 					//if msg is a task
// 						//Add task to queue
// 						//Choose elevator that will preform task
// 						//Send task to elevator
// 						//start timer for task
// 							//	-> if task not completed before the timer runs out ->dicconect elevator and redistribute the task
// 						//Update elevator overview
// 					//if msg is "done"
// 						//remove the ask from queue and update the elevator in overview
// 					//if msg is "stop"
// 						//update elevator in overwiev and redistribute the task
// 					//if msg is floor
// 						//update floor of elevator in overwiev.
// 				}
// 			}()
// 			log.Fatal(server.Start())

// 			//Broadcast of queue
// 			master, err := net.DialUDP("udp", nil, udpAddr)
// 			if err != nil {
// 				fmt.Println("Something went wrong: ", err)
// 				return
// 			}

// 			go func() {
// 				for {
// 					//Every x second send queue to backup
// 					time.Sleep(1 * time.Second)
// 					master.Write([]byte("test"))
// 					//should listen to the backup
// 					//setup the SetReadDeadline
// 				}
// 			}()

// 		case "backup":
// 			fmt.Print("Backup loop")

// 			//Brodcast msg on udpBackupPort
// 			//Listen for msg from master

// 			listner, err := net.ListenUDP("udp", udpAddr)
// 			if err != nil {
// 				fmt.Println("Something went wrong: ", err)
// 				return
// 			}

// 			for {
// 				buffer := make([]byte, 1024)
// 				listner.SetReadDeadline(time.Now().Add(2 * time.Second))
// 				_, _, err := listner.ReadFromUDP(buffer)
// 				if err != nil {
// 					//Chnage to backup to master when master is not sending within deadline
// 					listner.Close()
// 					state = "master"
// 					break
// 				} else {
// 					//Else save the received queue
// 					//BackupQueue =
// 					//return the message to master
// 				}
// 			}
// 		}
// 	}
// }