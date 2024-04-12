package elevcons

const N_Floors int = 4
const N_Buttons int = 3

const (
	MsgPort    = 10101
	UpdatePort = 32434
	TcpPort    = ":14068"
)

type ElevatorMode int

const (
	Primary        ElevatorMode = 0
	Backup         ElevatorMode = 1
	SingleElevator ElevatorMode = 2
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ElevatorBehaviour int

const (
	Idle      ElevatorBehaviour = 0
	Door_open ElevatorBehaviour = 1
	Moving    ElevatorBehaviour = 2
)

type ReqType int

const (
	NewReq       int = 0
	CompletedReq int = 1
)

type ReqStatus int

const (
	TurnOffLight int = 0
	TurnOnLight  int = 1
	TakeReq      int = 2
)

type ReturnStatus int

const (
	NoInfoNeeded  int = 0
	NeedAllInfo   int = 1
	NeedLightInfo int = 2
)
