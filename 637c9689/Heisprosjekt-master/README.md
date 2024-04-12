# Elevatorproject 

To be able to run the elevator, make sure that the terminal is openend in the `Heisprosjekt` directory. The `hall_request_assigner` executeable file is included in the `hra` folder, and it can be used with a Linux os. To run the code using another operating system, an equivalent `hall_request_assigner` executable file must downloaded from the [Project resources](https://github.com/TTK4145) Github repository.

When connecting to the elevator server, make sure to run the server with the same port as is used in the `elevio.Init()` function in the `main.go` file. To customize the elevator system, the number of floors can be adjusted in the `elevcons.go` file. Here you can also change the port numbers for both the TCP and UDP communication setup, these are called `TcpPort` for the TCP network communication and `MsgPort`and `UpdatePort` for UDP network communication.

To run the program make sure that the elevatorserver is ran first. This can be done writing `elevatorserver --port XXXXX` or `simelevatorserver --port XXXXX` in the terminal. Then, in a separate terminal write `go run main.go`. 


## Elevator Project Resources

* `hall_request_assigner` has been downloaded from the [Project resources](https://github.com/TTK4145/Project-resources/tree/master/cost_fns) repository
  
* `./network` has been downloaded from the [TTK4145-Network](https://github.com/TTK4145/Network-go) repository

* `./driver-go-master` has been downloaded from the [TTK4145-Network](https://github.com/TTK4145/driver-go) repository
