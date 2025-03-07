package controller

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	elevator "Driver-go/elevator/types"
	localController "Driver-go/elevatorController/controller/localController"
	"Driver-go/elevatorController/timer"
	"fmt"
	"os"
)

func elevatorMain() {
	// Initialize the driver connection to the elevator server.
	addr := os.Args[1]
	addr = "localhost:" + addr
	driver.Init(addr, elevator.N_FLOORS)

	// Create channels for receiving events from the driver.
	drv_buttons := make(chan elevator.ButtonEvent) // Button press events.
	drv_floors := make(chan int)                   // Floor sensor readings.
	drv_obstr := make(chan bool)                   // Obstruction switch events.
	drv_stop := make(chan bool)                    // Stop button press events.

	// Start goroutines to poll various elevator inputs continuously.
	go driver.PollButtons(drv_buttons)
	go driver.PollFloorSensor(drv_floors)
	go driver.PollObstructionSwitch(drv_obstr)
	go driver.PollStopButton(drv_stop)

	// Initialize the local controller.

	// Create custom timers for door and mobility events.
	doorTimer := timer.NewTimer()
	mobilityTimer := timer.NewTimer()
	var doorTimeoutCh, mobilityTimeoutCh <-chan bool

	// Infinite loop to process elevator events.
	for {
		select {
		// Button press events.
		case btnEvent := <-drv_buttons:
			// Process the button press event via the FSM.
			localController.OnRequestButtonPress(btnEvent.Floor, btnEvent.Button)
			fmt.Printf("Button Event: %+v\n", btnEvent)
			// If the door is open, (re)start the door timer.
			if localController.IsDoorOpen() {
				doorTimeoutCh = startTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
			}

		// Floor sensor events.
		case floor := <-drv_floors:
			fmt.Printf("Floor Arrival: %d\n", floor)
			localController.OnFloorArrival(e, floor)
			// When moving, restart the mobility timer on floor arrival.
			if localController.IsMoving() {
				mobilityTimeoutCh = startTimerChannel(mobilityTimer, types.MOBILITY_TIMEOUT_SEC)
			}

		// Obstruction events.
		case obstructed := <-drv_obstr:
			fmt.Printf("Obstruction Event: %+v\n", obstructed)
			localController.OnObstruction(e, obstructed)
			// If an obstruction is detected, stop the door timer.
			if obstructed {
				doorTimer.Stop()
				doorTimeoutCh = nil
			} else {
				// If the door is still open and the obstruction is cleared, restart the door timer.
				if localController.IsDoorOpen() {
					doorTimeoutCh = startTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
				}
			}

		// Door timeout event.
		case <-doorTimeoutCh:
			doorTimeoutCh = nil
			doorTimer.Stop()
			fmt.Println("Door timer expired")
			localController.OnDoorTimeout()
			// If the FSM now transitions to MOVING, start the mobility timer.
			if localController.IsMoving() {
				mobilityTimeoutCh = startTimerChannel(mobilityTimer, types.MOBILITY_TIMEOUT_SEC)
			}

		// Mobility timeout event.
		case <-mobilityTimeoutCh:
			mobilityTimeoutCh = nil
			mobilityTimer.Stop()
			fmt.Println("Mobility timer expired")
			localController.OnMobilityTimeout(e)
		}
	}
}
