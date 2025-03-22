package requests

import (
	"Exercise3/types"
)

type Dirn_behaviour_pair struct {
	Dp_dirn      types.Dirn
	Dp_behaviour types.Elevator_behaviour
}

func Requests_above(e types.Elevator) bool {
	//checking of there are any requests above current floor
	for floor := e.E_floor + 1; floor < types.N_floors; floor++ {
		//if there are any requests that means that one of the three buttons has been pressed
		for btn := 0; btn < types.N_buttons; btn++ {
			if e.E_requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_below(e types.Elevator) bool {
	for floor := 0; floor < e.E_floor; floor++ {
		for btn := 0; btn < types.N_buttons; btn++ {
			if e.E_requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_here(e types.Elevator) bool {
	for btn := 0; btn < types.N_buttons; btn++ {
		if e.E_requests[e.E_floor][btn] {
			return true
		}
	}
	return false
}

// a function to choose what direction the elevator should keep going in
func Requests_chooseDirn(e types.Elevator) Dirn_behaviour_pair {
	switch e.E_dirn {
	case types.Dirn_UP:
		//if its on the way up, it should continue if there is more requests above
		if Requests_above(e) {
			return Dirn_behaviour_pair{types.Dirn_UP, types.EB_Moving}
		} else if Requests_here(e) {
			return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_DoorOpen}
		} else if Requests_below(e) {
			return Dirn_behaviour_pair{types.Dirn_Down, types.EB_Moving}
		}

	case types.Dirn_Down:
		if Requests_below(e) {
			return Dirn_behaviour_pair{types.Dirn_Down, types.EB_Moving}
		} else if Requests_here(e) {
			return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_DoorOpen}
		} else if Requests_above(e) {
			return Dirn_behaviour_pair{types.Dirn_UP, types.EB_Moving}
		}
	case types.Dirn_Stop:
		if Requests_here(e) {
			return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_DoorOpen}
		} else if Requests_above(e) {
			return Dirn_behaviour_pair{types.Dirn_UP, types.EB_Moving}
		} else if Requests_below(e) {
			return Dirn_behaviour_pair{types.Dirn_Down, types.EB_Moving}
		}
	default:
		//the standard is just standing and waiting
		return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_idle}
	}
	return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_idle}
}

// should the elevator stop at any floors
func Requests_shouldStop(e types.Elevator) bool {
	switch e.E_dirn {
	case types.Dirn_Down:
		//should stop if there is a hall or cab call or just no requests below
		//since it's moving down it should only stop for hallcall down if there are requests below
		return e.E_requests[e.E_floor][types.B_HallDown] || e.E_requests[e.E_floor][types.B_Cab] || !Requests_below(e)
	case types.Dirn_UP:
		return e.E_requests[e.E_floor][types.B_HallUp] || e.E_requests[e.E_floor][types.B_Cab] || !Requests_above(e)

	default:
		return true
	}
}

// should all requests be cleared, and immediately, need info about elev and button
func Requests_shouldClearImmediately(e types.Elevator, btn_floor int, btn_type types.Button_type) bool {
	switch e.E_config.ClearRequestVariant {
	//everyone on the floor enters the elevator despite it going in the wrong direction
	case types.CV_All:
		return e.E_floor == btn_floor

	case types.CV_InDirn:
		//if only those going in the given direcation enters the elevator
		return e.E_floor == btn_floor && ((e.E_dirn == types.Dirn_Down && btn_type == types.B_HallDown) ||
			(e.E_dirn == types.Dirn_UP && btn_type == types.B_HallUp) ||
			e.E_dirn == types.Dirn_Stop || btn_type == types.B_Cab)

	default:
		return false
	}
}

func Requests_clearAtCurrentFloor(e types.Elevator) types.Elevator {
	//this one cleans the matrix
	switch e.E_config.ClearRequestVariant {
	case types.CV_All:
		//if everyone enters the elevator, we need to iterate through the buttons and clear them
		for btn := 0; btn < types.N_buttons; btn++ {
			e.E_requests[e.E_floor][btn] = false

		}

	case types.CV_InDirn:
		//clears the request of cab calls on the current floor
		e.E_requests[e.E_floor][types.B_Cab] = false
		switch e.E_dirn {
		case types.Dirn_Down:
			if !Requests_below(e) && !e.E_requests[e.E_floor][types.B_HallDown] {
				e.E_requests[e.E_floor][types.B_HallUp] = false
			}
			e.E_requests[e.E_floor][types.B_HallDown] = false

		case types.Dirn_UP:
			if !Requests_above(e) && !e.E_requests[e.E_floor][types.B_HallUp] {
				e.E_requests[e.E_floor][types.B_HallDown] = false
			}
			e.E_requests[e.E_floor][types.B_HallUp] = false

		case types.Dirn_Stop:
			e.E_requests[e.E_floor][types.B_HallUp] = false
			e.E_requests[e.E_floor][types.B_HallDown] = false

		default:
			e.E_requests[e.E_floor][types.B_HallUp] = false
			e.E_requests[e.E_floor][types.B_HallDown] = false
		}
	}

	return e
}
