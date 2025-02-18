package main

import (
	"fmt"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":30000")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(1)

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(2)
	//defer conn.Close()
	//buf := make([]byte, 1024)
	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		fmt.Println(3)
		if err != nil {
			fmt.Println(err)

			return
		}
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

	}
}
