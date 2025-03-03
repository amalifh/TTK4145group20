package fsm

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	"Driver-go/elevatorController/timer"
)

// Initialize the elevator state
var elevatorState = types.initElevator()

// Set all button lamps based on the elevator's current request state
func setAllLights(e types.Elevator) {
	for floor := 0; floor < types.N_FLOORS; floor++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			driver.SetButtonLamp(types.ButtonType(btn), floor, e.Requests[floor][btn])
		}
	}
}

// Initialize the elevator to move down when between floors
func FsmOnInitBetweenFloors() {
	driver.SetMotorDirection(types.MD_Down)   // Set motor to move down
	elevatorState.Dirn = types.MD_Down        // Set direction to down
	elevatorState.Behaviour = types.EB_Moving // Set elevator state to moving
}

// Handle a button press event, updating the elevator state accordingly
func FsmOnRequestButtonPress(btnFloor int, btnType types.ButtonType) {
	if elevatorState.ObstructionDetected || elevatorState.StopButtonPressed { // Does nothing if obstruction is detected
		return
	}

	switch elevatorState.Behaviour {
	case types.EB_DoorOpen:
		// If the elevator doors are open and a button is pressed, decide whether to start a timer
		if requests.RequestsShouldClearImmediately(elevatorState, btnFloor, btnType) {
			timer.TimerStart(elevatorState.Config.DoorOpenDuration_s) // Start door timer
		} else {
			// Otherwise, keep the request active but don't change the state
			elevatorState.Requests[btnFloor][btnType] = true
		}

	case types.EB_Moving:
		// If the elevator is already moving, simply mark the request as true
		elevatorState.Requests[btnFloor][btnType] = true

	case types.EB_Idle:
		// If the elevator is idle, determine the next movement direction and update the state
		elevatorState.Requests[btnFloor][btnType] = true
		pair := requests.RequestsChooseDirection(elevatorState) // Choose direction based on requests
		elevatorState.Dirn = pair.Dirn
		elevatorState.Behaviour = pair.Behaviour

		// Update behavior depending on the direction
		switch pair.Behaviour {
		case types.EB_DoorOpen:
			driver.SetDoorOpenLamp(true)                                        // Open the doors if we stop
			timer.TimerStart(elevatorState.Config.DoorOpenDuration_s)           // Start the door open timer
			elevatorState = requests.RequestsClearAtCurrentFloor(elevatorState) // Clear requests at current floor

		case types.EB_Moving:
			driver.SetMotorDirection(elevatorState.Dirn) // Set the motor direction to the chosen direction
		}
	}

	// Update all the lights to reflect the current request state
	setAllLights(elevatorState)
}

// Handle the elevator arriving at a floor
func FsmOnFloorArrival(newFloor int) {
	// Update the current floor and set the floor indicator
	elevatorState.Floor = newFloor
	driver.SetFloorIndicator(elevatorState.Floor)

	// If the elevator is moving, check if we should stop at the current floor
	switch elevatorState.Behaviour {
	case types.EB_Moving:
		if requests.RequestsShouldStop(elevatorState) { // Check if there's a request at the current floor
			// Stop the elevator and open the doors
			driver.SetMotorDirection(types.D_Stop)
			driver.SetDoorOpenLamp(true)
			elevatorState = requests.RequestsClearAtCurrentFloor(elevatorState) // Clear requests at current floor
			timer.TimerStart(elevatorState.Config.DoorOpenDuration_s)           // Start the door open timer
			setAllLights(elevatorState)
			elevatorState.Behaviour = types.EB_DoorOpen // Change state to door open
		}
	}
}

func FsmOnObstruction(obstructed bool) {
	if obstructed {
		elevatorState.ObstructionDetected = true
		switch elevatorState.Behaviour {
		case types.EB_DoorOpen:
			driver.SetMotorDirection(types.D_Stop)
			driver.SetDoorOpenLamp(true)
			setAllLights(elevatorState)
		}
	} else {
		if elevatorState.ObstructionDetected {
			setAllLights(elevatorState)
			elevatorState.ObstructionDetected = false
			if elevatorState.Behaviour == types.EB_Moving {
				driver.SetMotorDirection(elevatorState.Dirn)
			} else {
				driver.SetDoorOpenLamp(true)
				timer.TimerStart(elevatorState.Config.DoorOpenDuration_s)
			}
		}
	}
}

func FsmOnStop() {
	if !elevatorState.StopButtonPressed {
		elevatorState.StopButtonPressed = true
		driver.SetMotorDirection(types.D_Stop)
		driver.SetStopLamp(true)
	} else {
		elevatorState.StopButtonPressed = false
		driver.SetStopLamp(false)
		if elevatorState.ObstructionDetected {
			return
		}
		driver.SetMotorDirection(elevatorState.Dirn)
	}
}

// Handle door timeout, managing transitions based on the elevator's state
func FsmOnDoorTimeout() {
	switch elevatorState.Behaviour {
	case types.EB_DoorOpen:
		if elevatorState.ObstructionDetected || elevatorState.StopButtonPressed { // Does nothing if obstruction is detected
			return
		}
		// After the door timeout, determine whether to continue moving or remain idle
		pair := requests.RequestsChooseDirection(elevatorState) // Choose next direction based on requests
		elevatorState.Dirn = pair.Dirn
		elevatorState.Behaviour = pair.Behaviour

		// Update behavior based on direction
		switch elevatorState.Behaviour {
		case types.EB_DoorOpen:
			timer.TimerStart(elevatorState.Config.DoorOpenDuration_s)           // Start the door open timer again
			elevatorState = requests.RequestsClearAtCurrentFloor(elevatorState) // Clear requests at current floor
			setAllLights(elevatorState)                                         // Update lights

		case types.EB_Moving, types.EB_Idle:
			// If moving or idle, close the doors and start moving in the chosen direction
			driver.SetDoorOpenLamp(false)
			driver.SetMotorDirection(elevatorState.Dirn) // Move in the chosen direction
		}
	}
}
