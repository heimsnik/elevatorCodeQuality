package tcpnetwork

import  "time"

const NEW_ELEVATOR_CONNECTION_PORT string = "9999"

const BACKUP_PORT = "30000"

const ELEVATOR_PORT_OFFSET int = 10000

const IM_ALIVE_SIGNAL_MS_TIMEOUT time.Duration = 1500*time.Millisecond


const _tcpWriteDeadline = 100*time.Millisecond

const _bufferSize = 1024

const _checkSocketAliveInterval = 40 * time.Millisecond

const _primaryAcceptConnectionTimeout = 5 * time.Second

const _afterDeleteConnectionSleepTime = 5 * time.Second


