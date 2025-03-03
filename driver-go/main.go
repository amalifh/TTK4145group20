package main

import (
	"Driver-go/elevator/driver"
	elevator "Driver-go/elevator/types"
	"Driver-go/elevatorController/fsm"
	"Driver-go/elevatorController/timer"
	"fmt"
	"os"
	"time"
)

func main() {
	// Initialize the driver connection to the elevator server
	addr := os.Args[1]
	addr = "localhost:" + addr
	driver.Init(addr, elevator.N_FLOORS)
	// Handle the initliazitation in the FSM
	fsm.FsmOnInitBetweenFloors()

	// Create channels for receiving events from the driver
	drv_buttons := make(chan elevator.ButtonEvent) // For button presses
	drv_floors := make(chan int)                   // For floor sensor readings
	drv_obstr := make(chan bool)                   // For obstruction switch events
	drv_stop := make(chan bool)                    // For stop button press events

	// Start goroutines to poll various elevator inputs continuously
	go driver.PollButtons(drv_buttons)
	go driver.PollFloorSensor(drv_floors)
	go driver.PollObstructionSwitch(drv_obstr)
	go driver.PollStopButton(drv_stop)

	// Infinite loop to process elevator events
	for {
		select {
		// Handle button press events
		case a := <-drv_buttons:
			// Pass the button press event to the FSM for processing
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
			// Print the button press event for debugging purposes
			fmt.Printf("%+v\n", a)

		// Handle floor arrival events
		case a := <-drv_floors:
			// Print the floor arrival for debugging
			fmt.Printf("%+v\n", a)
			// Handle the floor arrival in the FSM
			fsm.FsmOnFloorArrival(a)

		// Handle obstruction events
		case a := <-drv_obstr:
			// Print the obstruction event if it occurred
			fmt.Printf("%+v\n", a)
			// Process the obstruction event in the FSM
			//fsm.FsmOnObstruction(a)


		// Periodic check every second for timer expiration
		case <-time.After(1 * time.Second):
			// If the timer has expired
			if timer.TimerTimedOut() {
				// Stop the timer and process the door timeout logic in the FSM
				timer.TimerStop()
				fsm.FsmOnDoorTimeout()
			}
		}
	}
}
