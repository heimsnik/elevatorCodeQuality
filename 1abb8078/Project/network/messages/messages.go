package messages

import (
	"Project/elevalgo"
	"Project/elevio"
	"encoding/json"
	"fmt"
	"reflect"
	"bytes"
	"strings"
)

// Primary <- everyone 
type M_Connected struct {
}

// Primary -> Backup
type M_BackupHallRequest struct {
	Data elevio.ButtonEvent
}
type M_BackupCabRequest struct {
	Id  int
	Data elevio.ButtonEvent
}
type M_DeleteHallRequest struct {
	Data elevio.ButtonEvent
}
type M_DeleteCabRequest struct {
	Id  int
	Data elevio.ButtonEvent
}
type M_PrimaryAlive struct {
	Data map[int]elevalgo.Elevator
}

// Primary <- Backup
type M_AckBackupHallRequest struct {
	Data elevio.ButtonEvent
}
type M_AckBackupCabRequest struct {
	Id  int
	Data elevio.ButtonEvent
}
type M_BackupAlive struct{}

// Primary -> ClientElevator
type M_DoRequest struct {
	Id int 
	Data elevio.ButtonEvent
}
type M_HallLights struct {
	Data [elevio.NUM_FLOORS][elevio.NUM_BUTTONS - 1]bool
}
type M_SpawnBackup struct{
	Id int 
}
type M_KILL struct{
	Id int
}

// Primary <- ClientElevator
type M_NewRequest struct {
	Data elevio.ButtonEvent
}
type M_CompletedRequest struct {
	Data elevio.ButtonEvent
}
type M_ElevatorAlive struct {
	Data elevalgo.Elevator
}

func MessageToBytes(message interface{}) []byte {
	var id int
	switch message.(type) {
	case M_BackupCabRequest:
		id = message.(M_BackupCabRequest).Id
	case M_DeleteCabRequest:
		id = message.(M_DeleteCabRequest).Id
	case M_AckBackupCabRequest:
		id = message.(M_AckBackupCabRequest).Id
	case M_DoRequest:
		id = message.(M_DoRequest).Id
	case M_SpawnBackup:
		id = message.(M_SpawnBackup).Id
	case M_KILL:
		id = message.(M_KILL).Id
	case M_HallLights:
		id = 255
	default:
		id = -1
	}
	
	var typ string = reflect.TypeOf(message).Name()

	dataString := dataToString(message)
	
	commonMessage := commonMessageFormat{id, typ, dataString}
	
	bytes, err := json.Marshal(commonMessage)
	if err != nil {
		fmt.Println("Error: json.Marshal failed (messageToBytes)", err)
		return []byte{}
	}
	return bytes
}

func BytesToMessage(bytes []byte) interface{} {
	// unpack byte array
	var commonMessage commonMessageFormat
	err := json.Unmarshal(bytes, &commonMessage)
	if err != nil {
		fmt.Println("Error: json.Unmarshal failed (bytesToMessage) error:",err)
		return nil
	}
	
	return  commonFormatToSpecificMessage(commonMessage)
}

func SplitMessages(data []byte) []string{
    strInput := string(data)
    strInput = strings.ReplaceAll(strInput, "}{", "}~{")
    byteSlices := bytes.Split([]byte(strInput), []byte("~"))

    // Convert byte slices to an array of strings
    result := make([]string, len(byteSlices))
    for i, part := range byteSlices {
        result[i] = string(part)
    }

    return result
}

func GetID(bytes []byte) int { 
	commonMessage := commonMessageFormat{}
	err := json.Unmarshal(bytes, &commonMessage)
	if err != nil {
		fmt.Println("Error: json.Unmarshal failed (getID)")
		return -1
	}
	return commonMessage.Id
}

func SetID(bytes []byte, length int, ID int) []byte {
	commonMessage := commonMessageFormat{}
	err := json.Unmarshal(bytes[:length], &commonMessage)
	if err != nil {
		fmt.Println("Error: json.Unmarshal failed (setID)")
		return []byte{}
	}
	commonMessage.Id = ID
	bytes, err = json.Marshal(commonMessage)
	if err != nil {
		fmt.Println("Error: json.Marshal failed (setID)")
		return []byte{}
	}
	return bytes
}

// id will be 255 to "broadcast" to all clients (MT_HallLights)
// id will be -1 if id is irrelevant
// data will be "" if data is irrelevant
type commonMessageFormat struct {
	Id   int
	Typ  string // type of message ("M_Request" etc)
	Data string
}

func dataToString(message interface{}) string {
	switch message.(type) {
	case M_BackupHallRequest:
		btnEvent := message.(M_BackupHallRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_BackupCabRequest:
		btnEvent := message.(M_BackupCabRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_DeleteHallRequest:
		btnEvent := message.(M_DeleteHallRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_DeleteCabRequest:
		btnEvent := message.(M_DeleteCabRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_PrimaryAlive:
		elevMap := message.(M_PrimaryAlive).Data
		elevMapBytes, err := json.Marshal(elevMap)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return string(elevMapBytes)
	case M_AckBackupHallRequest:
		btnEvent := message.(M_AckBackupHallRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_AckBackupCabRequest:
		btnEvent := message.(M_AckBackupCabRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_DoRequest:
		btnEvent := message.(M_DoRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_HallLights:
		hallLights := message.(M_HallLights).Data
		hallLightsBytes, err := json.Marshal(hallLights)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return string(hallLightsBytes)
	case M_NewRequest:
		btnEvent := message.(M_NewRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_CompletedRequest:
		btnEvent := message.(M_CompletedRequest).Data
		return fmt.Sprintf("%d,%d", btnEvent.Floor, btnEvent.Button)
	case M_ElevatorAlive:
		elev := message.(M_ElevatorAlive).Data
		elevBytes, err := json.Marshal(elev)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return string(elevBytes)

	default:
		return ""
	}
}

func commonFormatToSpecificMessage(commonMessage commonMessageFormat) interface{} {
	switch commonMessage.Typ {
	case "M_Connected":
		return M_Connected{}
	case "M_BackupHallRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_BackupHallRequest{Data: be}
	case "M_BackupCabRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_BackupCabRequest{Id: commonMessage.Id, Data: be}
	case "M_DeleteHallRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_DeleteHallRequest{Data: be}
	case "M_DeleteCabRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_DeleteCabRequest{Id: commonMessage.Id, Data: be}
	case "M_PrimaryAlive":
		var elevMap map[int]elevalgo.Elevator
		err := json.Unmarshal([]byte(commonMessage.Data), &elevMap)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return M_PrimaryAlive{Data: elevMap}
	case "M_AckBackupHallRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_AckBackupHallRequest{Data: be}
	case "M_AckBackupCabRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_AckBackupCabRequest{Id: commonMessage.Id, Data: be}
	case "M_BackupAlive":
		return M_BackupAlive{}
	case "M_DoRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_DoRequest{Id: commonMessage.Id, Data: be}
	case "M_HallLights":
		var hallLightMatrix [elevio.NUM_FLOORS][elevio.NUM_BUTTONS - 1]bool
		err := json.Unmarshal([]byte(commonMessage.Data), &hallLightMatrix)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return M_HallLights{Data: hallLightMatrix}
	case "M_SpawnBackup":
		return M_SpawnBackup{Id: commonMessage.Id}
	case "M_KILL":
		return M_KILL{Id: commonMessage.Id}
	case "M_NewRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_NewRequest{Data: be}
	case "M_CompletedRequest":
		var floor, button int
		fmt.Sscanf(commonMessage.Data, "%d,%d", &floor, &button)
		be := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
		return M_CompletedRequest{Data: be}
	case "M_ElevatorAlive":
		var elev elevalgo.Elevator
		err := json.Unmarshal([]byte(commonMessage.Data), &elev)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return M_ElevatorAlive{Data: elev}
	default:
		fmt.Println("toMessage(), Error: unknown message type")
		return nil
	}
}
