package udpnetwork

import "time"

const UDP_BROADCAST_PORT string = "16969"

const PRIMARY_ONLY_BROADCAST_PORT = "12333"

const UDP_BROADCAST_IP string = "255.255.255.255"


const _udpReadTimeout = 6 * time.Second

const _bufferSize = 1024

const _broadcastInterval = 2 * time.Second

const _labIPPrefix = "10"

const _randomIPandPort = "10.10.10.100:100"