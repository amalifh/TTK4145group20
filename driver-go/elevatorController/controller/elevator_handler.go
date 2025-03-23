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

			if localCtrl.CurrentElevator.Behaviour == types.EB_Moving {
				driver.SetMotorDirection(localCtrl.DirectionConverter(localCtrl.CurrentElevator.Direction))
			}

			if IsDoorOpen() {
				doorTimeoutCh = StartTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
			}

		// Floor sensor events.
		case floor := <-drv_floors:
			driver.SetFloorIndicator(floor)
			fmt.Printf("Floor Arrival: %d\n", floor)
			localCtrl.OnFloorArrival(floor)
			if IsMoving() {
				mobilityTimeoutCh = StartTimerChannel(mobilityTimer, types.MOBILITY_TIMEOUT_SEC)
			} else {
				driver.SetMotorDirection(types.MD_Stop)
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
				// driver.SetDoorOpenLamp(false)
			}

		// Door timeout event.
		case <-doorTimeoutCh:
			doorTimeoutCh = nil
			doorTimer.Stop()
			fmt.Println("Door timer expired")
			driver.SetDoorOpenLamp(false)

			localCtrl.OnDoorTimeout()

			mobilityTimer.Stop()
			mobilityTimeoutCh = nil

			if IsMoving() {
				newDir := localCtrl.DirectionConverter(GetDirection())
				driver.SetMotorDirection(newDir)
				fmt.Printf("New direction: %v\n", newDir)
				// Optionally, restart the mobility timer.
				mobilityTimeoutCh = StartTimerChannel(mobilityTimer, types.MOBILITY_TIMEOUT_SEC)
			} else {
				fmt.Printf("No pending requests!")
				driver.SetMotorDirection(types.MD_Stop)
			}

		// Mobility timeout event.
		case <-mobilityTimeoutCh:
			mobilityTimeoutCh = nil
			mobilityTimer.Stop()
			fmt.Println("Mobility timer expired")
			localCtrl.OnMobilityTimeout()
			driver.SetMotorDirection(types.MD_Stop) //How will it detect it can move again if it is set to stop?
		}
	}
}
