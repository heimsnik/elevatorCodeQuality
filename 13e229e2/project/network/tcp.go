package network

import(
	"fmt"
	"net"
	"bufio"
	"strings"
	"strconv"
)

func HandleConnection(conn net.Conn) {
    // fmt.Printf("Serving %s\n", conn.RemoteAddr().String())
    for {
            netData, err := bufio.NewReader(conn).ReadString('\n')
            if err != nil {
                    fmt.Println(err)
                    return
            }

            temp := strings.TrimSpace(string(netData))
            if temp == "STOP" {
                    break
            }

            result := strconv.Itoa(69) + "\n"
            conn.Write([]byte(string(result)))
    }
    conn.Close()
}

func SetupTCPServer(port string) {
	// port needs to have ":"-prefix
	if !(strings.Contains(port, ":")) {
		port = ":" + port
	}
    ln, err := net.Listen("tcp", port)
    if err != nil {
            fmt.Println(err)
            return
    }
    defer ln.Close()

    for {   
            conn, err := ln.Accept()
            if err != nil {
                    fmt.Println(err)
                    return
            }
            go HandleConnection(conn)
    }
}

// func send(conn *net.TCPConn, port string, msg /*insert struct type here*/ ) {
// 	// adr, err := net.ResolveTCPAddr("tcp", "10.24.32.136:20000")
// 	// ln, err := net.DialTCP("tcp",nil, adr)
// 	// if err != nil {
// 	// 	fmt.Printf("Something went wrong when creating socket: ", err)
// 	// 	return
// 	// }
	
// 		data := []byte("Connect to 10.24.35.255:42106\000")
// 		// addr := &net.UDPAddr{
// 		// 	IP: net.ParseIP("10.100.23.129"),
// 		// 	Port: 20012,
// 		// }



// 		conn.Write(data) //, addr)
// 		time.Sleep(5 * time.Second)
// 		conn.Close()
	
// }

func receive(conn *net.TCPConn) {
	
	buffer := make([]byte, 1024)
	
	for {	
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Something went wrong when reading from buffer: ", err)
			return
		}
		fmt.Printf("Received: %s \n ", string(buffer[:n]))
	}
	
	conn.Close()
	
}