package config

import "time"


const NumFloors = 4
const NumElevators = 3
const DoorOpenDuration = 3 * time.Second
const taskWatchdogDuraion = 2 * time.Second

const TcpPort = ":3000"
const UdpMasterPort =	":6000"
const UdpBackupPort =	":7000"
const BroadcastInterval =	500 * time.Millisecond
const ReconnectInterval = 10 * time.Second
