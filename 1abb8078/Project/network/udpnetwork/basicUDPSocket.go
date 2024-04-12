package udpnetwork


type BasicUDPSocket struct {
	In 	chan string
	Out chan string

	stop chan bool
}

func newBasicUDPSocket() *BasicUDPSocket {
	return &BasicUDPSocket{
		In: make(chan string),
		Out: make(chan string),
		stop: make(chan bool),
	}
}

func (socket *BasicUDPSocket) stopBasicSocket() {
	go func() {
		for{
			socket.stop <- true
		}
	}()
}