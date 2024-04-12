package tcpnetwork


type BasicTCPSocket struct {
	In  chan []byte
	Out chan []byte

	active    bool
	stop 	chan bool
}


func NewBasicTCPSocket() BasicTCPSocket {
	return BasicTCPSocket{
		In:        make(chan []byte),
		Out:       make(chan []byte),
		active:    false,
		stop: 		make(chan bool),
	}
}

func (socket *BasicTCPSocket) IsActive() bool {
	return socket.active
}

func (socket *BasicTCPSocket) stopBasicSocket() {
	socket.active = false
	go func(){
		for{
			socket.stop <- true
		}
	}()
}