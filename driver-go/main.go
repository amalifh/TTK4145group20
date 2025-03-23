package main

import (
	//"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"driver-go/controller"
	"driver-go/elevator/driver"
	. "driver-go/elevator/types"
	"driver-go/network/bcast"
	"driver-go/network/peers"
	"driver-go/requests"
	"driver-go/sync"
)

func main() {
	var (
		runType string
		id      string
		e       driver.Elev_type
		ID      int
		simPort string
	)

	/*
		flag.StringVar(&runType, "run", "", "run type")
		flag.StringVar(&id, "id", "0", "id of this peer")
		flag.StringVar(&simPort, "simPort", "44523", "simulation server port")
		flag.Parse()
		ID, _ = strconv.Atoi(id)

		if runType == "sim" {
			e = hw.ET_Simulation
			fmt.Println("Running in simulation mode!")
		}
	*/

	contollerChans := controller.fsmChannels{
		OrderComplete:  make(chan int),
		Elevator:       make(chan ElevInfo),
		NewOrder:       make(chan ButtonEvent),
		ArrivedAtFloor: make(chan int),
	}

	syncChans := sync.SyncChannels{
		UpdateGovernor:  make(chan [N_ELEVATORS]ElevInfo),
		UpdateSync:      make(chan ElevInfo),
		OrderUpdate:     make(chan Keypress),
		OnlineElevators: make(chan [N_ELEVATORS]bool),
		IncomingMsg:     make(chan Message),
		OutgoingMsg:     make(chan Message),
		PeerUpdate:      make(chan peers.PeerUpdate),
		PeerTxEnable:    make(chan bool),
	}
	var (
		btnsPressedChan = make(chan ButtonEvent)
		lightUpdateChan = make(chan [types.N_ELEVATORS]ElevInfo)
	)

	driver.Init("localhost:"+simPort, N_FLOORS)

	go driver.ButtonPoller(btnsPressedChan)
	go driver.FloorIndicatorLoop(esmChans.ArrivedAtFloor)
	go controller.RunElevator(esmChans)
	go requests.requestAssigner(ID, btnsPressedChan, lightUpdateChan, esmChans.OrderComplete, esmChans.NewOrder, esmChans.Elevator,
		syncChans.OrderUpdate, syncChans.UpdateSync, syncChans.UpdateGovernor, syncChans.OnlineElevators)
	go requests.lightsUpdater(lightUpdateChan, ID)
	go sync.Synchronise(syncChans, ID)
	go bcast.Transmitter(42034, syncChans.OutgoingMsg)
	go bcast.Receiver(42034, syncChans.IncomingMsg)
	go peers.Transmitter(42035, id, syncChans.PeerTxEnable)
	go peers.Receiver(42035, syncChans.PeerUpdate)
	go killSwitch()

	select {}
}

func killSwitch() {
	// killSwitch turns the motor off if the program is killed with CTRL+C.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	driver.SetMotorDirection(types.MD_Stop)
	fmt.Println("\x1b[31;1m", "User terminated program.", "\x1b[0m")
	for i := 0; i < 10; i++ {
		driver.SetMotorDirection(types.MD_Stop)
		if i%2 == 0 {
			driver.SetStopLamp(1)
		} else {
			driver.SetStopLamp(0)
		}
		time.Sleep(200 * time.Millisecond)
	}
	driver.SetMotorDirection(types.MD_Stop)
	os.Exit(1)
}
