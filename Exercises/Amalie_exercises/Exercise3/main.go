package main

import (
	"Exercise3/timer"
	"Exercise3/types"
	"Exercise3/requests"
	"Exercise3/fms"
	"fmt"
	"time"
	"os"
)

func main() {
	//connect to the elevator
	addr := os.Args[1]
	addr = "localhost: "+addr
	driver.init(addr, types.N_floors)
	fsm.fsm_initBetweenFloors()

	ch_button := make(chan types.Button_event)
	ch_floors := make(chan int)
	ch_obstr := make(chan bool)
	ch_stop := make(chan bool)

//creating goroutines for every poll process
	go driver.PollButtons(ch_button)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)

//in need of a infinite for loop to run forever
	for {
		select {
		case a := <- ch_button:
			fsm.fsm_requestButtonPress(a.Floor, a.Button)
		case b := <- ch_floors:
			fsm.fsm_floorArrival(b)
		case c := <- ch_obstr:
			fsm.fsm_obstruction(c)
		case d := <- ch_stop:
			if d {
				fsm.fsm_stop
			}
		case <-time.After(1 * time.Second):
			// If the timer has expired
			if timer.timerTimedOut() {
				// Stop the timer and process the door timeout logic in the FSM
				timer.timerStop()
				fsm.fsm_doorTimeout()
		}

	}
}
}