package controller

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	localCtrl "Driver-go/elevatorController/controller/localController"
	"Driver-go/elevatorController/timer"
	"fmt"
	"os"
)

func ElevatorMain() {
	// Check command-line argument for port.
	if len(os.Args) < 2 {
		fmt.Println("Usage: <program> <port>")
		return
	}
	addr := os.Args[1]
	addr = "localhost:" + addr
	driver.Init(addr, types.N_FLOORS)

	// Create channels for receiving events.
	drv_buttons := make(chan types.ButtonEvent) // Button press events.
	drv_floors := make(chan int)                // Floor sensor events.
	drv_obstr := make(chan bool)                // Obstruction switch events.
	drv_stop := make(chan bool)                 // Stop button events.

	// Start goroutines to poll elevator inputs.
	go driver.PollButtons(drv_buttons)
	go driver.PollFloorSensor(drv_floors)
	go driver.PollObstructionSwitch(drv_obstr)
	go driver.PollStopButton(drv_stop)

	// Create custom timers for door and mobility events.
	doorTimer := timer.NewTimer()
	mobilityTimer := timer.NewTimer()
	var doorTimeoutCh, mobilityTimeoutCh <-chan bool

	// Main event loop.
	for {
		select {
		// Button press events.
		case btnEvent := <-drv_buttons:
			fmt.Printf("Button Event: %+v\n", btnEvent)
			localCtrl.OnRequestButtonPress(btnEvent.Floor, int(btnEvent.Button))
			if IsDoorOpen() {
				doorTimeoutCh = StartTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
			}

		// Floor sensor events.
		case floor := <-drv_floors:
			fmt.Printf("Floor Arrival: %d\n", floor)
			localCtrl.OnFloorArrival(floor)
			if IsMoving() {
				mobilityTimeoutCh = StartTimerChannel(mobilityTimer, types.MOBILITY_TIMEOUT_SEC)
			}

		// Obstruction events.
		case obstructed := <-drv_obstr:
			fmt.Printf("Obstruction Event: %+v\n", obstructed)
			if localCtrl.OnObstruction(obstructed) {
				doorTimer.Stop()
				doorTimeoutCh = nil
			} else {
				if IsDoorOpen() {
					doorTimeoutCh = StartTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
				}
			}

		// Door timeout event.
		case <-doorTimeoutCh:
			doorTimeoutCh = nil
			doorTimer.Stop()
			fmt.Println("Door timer expired")
			localCtrl.OnDoorTimeout()
			if IsMoving() {
				driver.SetDoorOpenLamp(false)
				driver.SetMotorDirection(DirectionConverter(GetDirection()))
				mobilityTimeoutCh = StartTimerChannel(mobilityTimer, types.MOBILITY_TIMEOUT_SEC)
			} else if IsDoorOpen() {
				doorTimeoutCh = StartTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
			} else if IsIdle() {
				driver.SetDoorOpenLamp(false)
			}

		// Mobility timeout event.
		case <-mobilityTimeoutCh:
			mobilityTimeoutCh = nil
			mobilityTimer.Stop()
			fmt.Println("Mobility timer expired")
			localCtrl.OnMobilityTimeout()
			driver.SetMotorDirection(types.MD_Stop)
		}
	}
}
