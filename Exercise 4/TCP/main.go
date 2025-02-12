package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func Primary(lastValue int) {
	fmt.Println("Primary initializing")
	ln, _ := net.Listen("tcp", ":20022")
	//exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()            // Linux
	exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go").Run() // Windows
	fmt.Println("Primary listening")
	conn, _ := ln.Accept()
	fmt.Println("Primary connected")

	var str string
	for {
		lastValue++
		str = strconv.Itoa(int(lastValue))
		data := []byte(str)
		conn.Write(data)
		fmt.Println(lastValue)
		time.Sleep(1500 * time.Millisecond)
	}
}

func Backup() int {
	fmt.Println("Backup initializing")
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:20022")
	fmt.Println("Backup resolved")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("Error dialing")
		return 0
	}
	fmt.Println("Backup dialed")

	var num int
	for {
		buf := make([]byte, 1024)
		value, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Sending num:", num)
			conn.Close()
			return num
		} else {
			data := buf[0:value]
			strdata := string(data)
			num, _ = strconv.Atoi(strdata)
			fmt.Println(num)
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
