# Elevator project

## How to build
To compile the correct the executable file, run the command:
```bash
go build -o RunElevator main.go
```

We now need the executable for running the cost_function of the elevators. The code is retrieved from the github repository `https://github.com/TTK4145/Project-resources/tree/master/elev_algo`. To get the correct executable, perform the following steps:
1. Copy the repository using: `git clone https://github.com/TTK4145/Project-resources.git`
2. Build the content of the hall_request_assigner folder by running the `build.sh` file.
3. Copy the resulting executable `hall_request_assigner` into the same folder as the `RunElevator`-file.
    - It is important to note here that the two executables are in the same folder location, as the code will not work otherwise.   

## How to run
To run the elevator, while in the same folder as `RunElevator` execute the command: `./RunProgram.sh #id #port`, where #id is the id corresponding to the elevator (1, 2 or 3) and #port is a suitable port the elevator can use. 