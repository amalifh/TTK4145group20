package fsm

import (
	"Exercise3/driver"
	"Exercise3/requests"
	"Exercise3/timer"
	"Exercise3/types"
)

// starting of by just initializing the elevator
var elevState = types.Elevator_uninitialized()

// the whole point of this is to start the processes, having the functions actually working together!
func SetAllLights(e types.Elevator) {
	for floor := 0; floor < types.N_floors; floor++ {
		for btn := 0; btn < types.N_buttons; btn++ {
			driver.SetButtonLamp(types.Button_type(btn), floor, e.E_requests[floor][btn])
		}
	}
}

func Fsm_initBetweenFloors() {
	driver.SetMotorDirection(types.Dirn_Down)
	elevState.E_dirn = types.Dirn_Down
	elevState.E_behaviour = types.EB_Moving
}

// function to how the elevator should handle button presses
func Fsm_requestButtonPress(btnFloor int, btnType types.Button_type) {
	//if obstruction is on or stop is pressed the elevator is to do nothing
	if elevState.E_obstruction || elevState.E_stop {
		return
	}
	switch elevState.E_behaviour {
	case types.EB_DoorOpen:
		if requests.Requests_shouldClearImmediately(elevState, btnFloor, btnType) {
			timer.TimerStart(elevState.E_config.DoorOpen_s)
		} else {
			elevState.E_requests[btnFloor][btnType] = true
		}
	case types.EB_Moving:
		elevState.E_requests[btnFloor][btnType] = true

	case types.EB_idle:
		elevState.E_requests[btnFloor][btnType] = true
		pair := requests.Requests_chooseDirn(elevState)
		elevState.E_dirn = pair.Dp_dirn
		elevState.E_behaviour = pair.Dp_behaviour

		switch elevState.E_behaviour {
		case types.EB_DoorOpen:
			driver.SetDoorOpenLamp(true)
			timer.TimerStart(elevState.E_config.DoorOpen_s)
			elevState = requests.Requests_clearAtCurrentFloor(elevState)

		case types.EB_Moving:
			driver.SetMotorDirection(elevState.E_dirn)

		}

	}
	SetAllLights(elevState)
}

func Fsm_floorArrival(newFloor int) {
	elevState.E_floor = newFloor

	driver.SetFloorIndicator(elevState.E_floor)

	//wanting to check if elevator should stop at the new floor when moving
	switch elevState.E_behaviour {
	case types.EB_Moving:
		if requests.Requests_shouldStop(elevState) {
			driver.SetMotorDirection(types.Dirn_Stop)
			driver.SetDoorOpenLamp(true)
			elevState = requests.Requests_clearAtCurrentFloor(elevState)
			timer.TimerStart(elevState.E_config.DoorOpen_s)
			SetAllLights(elevState)
			elevState.E_behaviour = types.EB_DoorOpen
		}
	}
}

func Fsm_obstruction(obstructed bool) {
	if obstructed {
		elevState.E_obstruction = true
		switch elevState.E_behaviour {
		case types.EB_DoorOpen:
			driver.SetMotorDirection(types.Dirn_Stop)
			driver.SetDoorOpenLamp(true)
			SetAllLights(elevState)
		}
	} else {
		if elevState.E_obstruction {
			SetAllLights(elevState)
			elevState.E_obstruction = false
			if elevState.E_behaviour == types.EB_Moving {
				driver.SetMotorDirection(elevState.E_dirn)
			} else {
				driver.SetDoorOpenLamp(true)
				timer.TimerStart(elevState.E_config.DoorOpen_s)
			}
		}
	}
}

func Fsm_stop() {
	if !elevState.E_stop {
		elevState.E_stop = true
		driver.SetMotorDirection(types.Dirn_Stop)
		driver.SetStopLamp(true)
	} else {
		elevState.E_stop = false
		driver.SetStopLamp(false)
		if elevState.E_obstruction {
			return
		}
		driver.SetMotorDirection(elevState.E_dirn)
	}
}

func Fsm_doorTimeout() {
	switch elevState.E_behaviour {
	case types.EB_DoorOpen:
		pair := requests.Requests_chooseDirn(elevState)
		elevState.E_dirn = pair.Dp_dirn
		elevState.E_behaviour = pair.Dp_behaviour

		switch elevState.E_behaviour {
		case types.EB_DoorOpen:
			timer.TimerStart(elevState.E_config.DoorOpen_s)
			elevState = requests.Requests_clearAtCurrentFloor(elevState)
			SetAllLights(elevState)
		case types.EB_idle, types.EB_Moving:
			driver.SetDoorOpenLamp(false)
			driver.SetMotorDirection(elevState.E_dirn)

		}

	}
}
