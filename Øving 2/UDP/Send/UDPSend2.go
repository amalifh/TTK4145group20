package main

import (
	"fmt"
	"net"
	"time"
)

func UDPsend() {
	addr, _ := net.ResolveUDPAddr("udp", "10.100.23.204:20022")
	//fmt.Println("UDPSend debug")
	conn, _ := net.DialUDP("udp", nil, addr)
	conn.Write([]byte("It's working!"))
	time.Sleep(1500 * time.Millisecond)
}

func UDPreceive() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":20022")
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

func main() {

	go UDPsend()
	go UDPreceive()

	time.Sleep(5000 * time.Millisecond)

}
