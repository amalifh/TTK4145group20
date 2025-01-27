package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

	func main() {
		addr, _ := net.ResolveTCPAddr("tcp", "10.100.23.204:33546")
		conn, _ := net.DialTCP("tcp", nil, addr)
		conn.Write([]byte("It's working!\x00"))
		data,_ := bufio.NewReader(conn).ReadString('\x00')
		fmt.Println("Received: ", string(data))
		time.Sleep(1500 * time.Millisecond)
	}