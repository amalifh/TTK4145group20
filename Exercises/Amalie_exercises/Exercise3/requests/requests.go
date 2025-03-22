package request

import (
	"Exercise3/types"
	. "Exercise3/types"
)

type Dirn_behaviour_pair struct {
	dp_dirn   Dirn
	behaviour Elevator_behaviour
}

func requests_above(e types.Elevator) bool {
	//checking of there are any requests above current floor
	for floor = e.E_floor+1; floor < types.N_floors; floor++{
		//if there are any requests that means that one of the three buttons has been pressed
		for btn := 0; btn < types.N_buttons; btn++ {
			if e.E_requests[floor][btn] {
				return true
			}
		}
	}
	return false
}


func requests_below(e types.Elevator) bool {
	for floor := 0; floor < e.E_floor; floor++ {
		for btn := 0; btn < types.N_buttons; btn ++ {
			if e.E_requests[floor][btn]{
				return true
			}
		}
	} 
	return false
}

func requests_here(e types.Elevator) bool{
	for btn := 0; btn < types.N_buttons; btn++ {
		if e.E_requests[e.E_floor][btn] {
			return true
		}
	}
	return false
}

func requests_chooseDirn(e types.Elevator)
// think about what makes sense to check first, depending on the direction the elevator is moving in
func requests_chooseDirn(e types.Elevator) Dirn_behaviour_pair {
	switch e.E_dirn {
	case types.Dirn_UP:
		if requests_above(e) {
			return Dirn_behaviour_pair{types.Dirn_UP, types.EB_Moving}
		} else if requests_here(e) {
			return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_DoorOpen}
		} else if requests_below(e) {
			return Dirn_behaviour_pair{types.Dirn_Down, types.EB_Moving}
		}
	case types.Dirn_Down:
		if requests_below(e) {
			return Dirn_behaviour_pair{types.Dirn_UP, types.EB_Moving}
		} else if requests_here(e) {
			return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_DoorOpen}
		} else if requests_above(e) {
			return Dirn_behaviour_pair{types.Dirn_Down, types.EB_Moving}
		}
	case types.Dirn_Stop:
		if requests_here(e) {
			return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_DoorOpen}
		} else if requests_above(e) {
			return Dirn_behaviour_pair{types.Dirn_UP, types.EB_Moving}
		} else if requests_below(e) {
			return Dirn_behaviour_pair{types.Dirn_Down, types.EB_Moving}
		}

	}
	return Dirn_behaviour_pair{types.Dirn_Stop, types.EB_idle}
}

func requests_shouldStop(e types.Elevator) bool {
	switch e.E_dirn {
	case types.Dirn_Down:
		return e.E_requests[e.E_floor][types.B_HallDown] || e.E_requests[e.E_floor][types.B_Cab] || !requests_below(e)

	case types.Dirn_UP:
		return e.E_requests[e.E_floor][types.B_HallUp] || e.E_requests[e.E_floor][types.B_Cab] || !requests_above(e)

	default:
		return true
	}
}

func requests_shouldClearImmediately(e types.Elevator, btn_floor int, btn_type types.Button_type) bool {
	switch e.E_config.ClearRequestVariant {
	case types.CV_All:
		return e.E_floor == btn_floor

	case types.CV_InDirn:
		return e.E_floor == btn_floor && ((e.E_dirn == types.Dirn_UP && btn_type == types.B_HallDown) || (e.E_dirn == types.Dirn_Down && btn_type == types.B_HallDown) ||
			e.E_dirn == types.Dirn_Stop || btn_type == types.B_Cab)
	default:
		return false
	}
}

func requests_clearAtCurrentFloor(e types.Elevator) types.Elevator {
	switch 
}
