package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func Primary(i int) {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:20022")
	conn, _ := net.DialTCP("tcp", nil, addr)
	exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()

	for{
		i++
		str := strconv.Itoa(i)
		conn.Write([]byte(str))
		fmt.Println(i)
		time.Sleep(1500 * time.Millisecond)		
	}
}

func Backup() int {
	ln, _ := net.Listen("tcp", ":20022")
	conn, _ := ln.Accept()
	fmt.Println("Backup connected")

	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		_, err := conn.Read(buf)
		if err != nil{
			//fmt.Println(err)
			return 0
		}
	}
}


func main() {
	var wg sync.WaitGroup
	lastValue := 0


	wg.Add(1)
	lastValue = Backup()
	Primary(lastValue)
	wg.Wait()
	
	

}

