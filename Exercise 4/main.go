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
	addr, _ := net.ResolveUDPAddr("udp", "localhost:20022")
	//fmt.Println("UDPSend debug")
	conn, _ := net.DialUDP("udp", nil, addr)
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
	fmt.Println("Backup started")
	udpAddr, err := net.ResolveUDPAddr("udp", ":20022")
	if err != nil {
		return 0
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return 0 
	}
	fmt.Println("Backup connected")

	defer conn.Close()
	//buf := make([]byte, 1024)
	for {
		fmt.Println("Backup loop")
		buf := make([]byte, 1024)
		_,_, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
			return 0
		}
		TimerStart(3.25)
		if(TimerTimedOut()){
			lastValue, _:= strconv.Atoi(string(buf[len(buf)-1]))
			return lastValue
		}
		TimerStop()
		fmt.Println("Timer problem!")
	}
	return 0
}


var (
	timerEndTime time.Time
	timerActive  bool      
)

func TimerStart(duration float64) {
	timerEndTime = time.Now().Add(time.Duration(duration * float64(time.Second)))
	timerActive = true
}


func TimerStop() {
	timerActive = false
}

func TimerTimedOut() bool {
	return timerActive && time.Now().After(timerEndTime)
}


func main() {
	lastValue := 0
	//Create main shell and backup
	var wg sync.WaitGroup

	wg.Add(1)
	lastValue = Backup()
	Primary(lastValue)
	wg.Wait()
	
	
	//Primary broadcast that it is still alive
}

//Heartbeat to solve time to notice dead connection