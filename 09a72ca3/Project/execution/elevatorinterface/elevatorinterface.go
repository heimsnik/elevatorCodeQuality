// Adapted from https://github.com/TTK4145/driver-go

package elevatorinterface

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const PollRate = 20 * time.Millisecond

var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn

func Init(addr string) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func Read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func Write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func ToByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func ToBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
