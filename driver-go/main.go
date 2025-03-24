package main

import (
	//"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"Driver-go/controller"
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	"Driver-go/requests"
	"Driver-go/sync"
)

func main() {
	var (
		// e       driver.Elev_type
		ID      int
		addr string
		sID string
	)

	if len(os.Args) < 3 {
		fmt.Println("Usage: <program> <port> <id>")
		return
	}
	addr = os.Args[1]
	sID = os.Args[2]
	addr = "localhost:" + addr
	ID, _ = strconv.Atoi(sID)

	controllerChans := controller.FsmChannels{
		RequestsComplete:  make(chan int),
		Elevator:       make(chan types.ElevInfo),
		NewRequest:       make(chan types.ButtonEvent),
		ArrivedAtFloor: make(chan int),
	}

	syncChans := sync.SyncChannels{
		UpdateReqAssigner:  make(chan [types.N_ELEVATORS]types.ElevInfo),
		UpdateSync:      make(chan types.ElevInfo),
		RequestsUpdate:     make(chan types.ButtonEvent),
		AliveElevators: make(chan [types.N_ELEVATORS]bool),
		IncomingMsg:     make(chan types.NetworkMessage),
		OutgoingMsg:     make(chan types.NetworkMessage),
		PeerUpdate:      make(chan peers.PeerUpdate),
		PeerTxEnable:    make(chan bool),
	}
	var (
		btnsPressedChan = make(chan types.ButtonEvent)
		lightUpdateChan = make(chan [types.N_ELEVATORS]types.ElevInfo)
		obstructionChan = make(chan bool)
	)

	driver.Init(addr, types.N_FLOORS)

	go driver.PollButtons(btnsPressedChan)
	go driver.PollFloorSensor(controllerChans.ArrivedAtFloor)
	go driver.PollObstructionSwitch(obstructionChan)
	go controller.ElevatorHandler(controllerChans)
	go requests.RequestAssigner(ID, btnsPressedChan, lightUpdateChan, controllerChans.RequestsComplete, controllerChans.NewRequest, controllerChans.Elevator, 
		syncChans.RequestsUpdate, syncChans.UpdateSync, syncChans.UpdateReqAssigner, syncChans.AliveElevators)
	go requests.LightsUpdater(lightUpdateChan, ID)
	go sync.Synchronise(syncChans, ID)
	go bcast.Transmitter(42034, syncChans.OutgoingMsg)
	go bcast.Receiver(42034, syncChans.IncomingMsg)
	go peers.Transmitter(42035, sID, syncChans.PeerTxEnable)
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
			driver.SetStopLamp(true)
		} else {
			driver.SetStopLamp(false)
		}
		time.Sleep(200 * time.Millisecond)
	}
	driver.SetMotorDirection(types.MD_Stop)
	os.Exit(1)
}