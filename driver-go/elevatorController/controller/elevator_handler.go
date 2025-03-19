package controller

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	localCtrl "Driver-go/elevatorController/controller/localController"
	"Driver-go/elevatorController/timer"
	"fmt"
)

func ElevatorHandler(drv_buttons <-chan types.ButtonEvent, drv_floors <-chan int, drv_obstr <-chan bool) {
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
