package requests

import "Driver-go/elevator"

// Struct to hold the direction and behavior pair
type DirnBehaviourPair struct {
	Dirn      elevator.Dirn
	Behaviour elevator.ElevatorBehaviour
}

// Check if there are any requests for floors above the current floor
func RequestsAbove(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < elevator.N_FLOORS; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true // Return true if any request exists above
			}
		}
	}
	return false // No requests above
}

// Check if there are any requests for floors below the current floor
func RequestsBelow(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true // Return true if any request exists below
			}
		}
	}
	return false // No requests below
}

// Check if there are any requests for the current floor
func RequestsHere(e elevator.Elevator) bool {
	for btn := 0; btn < elevator.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true // Return true if any request exists at the current floor
		}
	}
	return false // No requests at the current floor
}

// Determine the next movement direction and behavior based on the requests
func RequestsChooseDirection(e elevator.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevator.D_Up:
		// If moving up, check if there are requests above
		if RequestsAbove(e) {
			return DirnBehaviourPair{elevator.D_Up, elevator.EB_Moving}
		} else if RequestsHere(e) {
			// If there are requests at the current floor, open doors
			return DirnBehaviourPair{elevator.D_Down, elevator.EB_DoorOpen}
		} else if RequestsBelow(e) {
			// If no requests above, but requests below, move down
			return DirnBehaviourPair{elevator.D_Down, elevator.EB_Moving}
		}
	case elevator.D_Down:
		// If moving down, check if there are requests below
		if RequestsBelow(e) {
			return DirnBehaviourPair{elevator.D_Down, elevator.EB_Moving}
		} else if RequestsHere(e) {
			// If there are requests at the current floor, open doors
			return DirnBehaviourPair{elevator.D_Up, elevator.EB_DoorOpen}
		} else if RequestsAbove(e) {
			// If no requests below, but requests above, move up
			return DirnBehaviourPair{elevator.D_Up, elevator.EB_Moving}
		}
	case elevator.D_Stop:
		// If stopped, check if there are requests at the current floor
		if RequestsHere(e) {
			return DirnBehaviourPair{elevator.D_Stop, elevator.EB_DoorOpen}
		} else if RequestsAbove(e) {
			// If no requests at the current floor, but requests above, move up
			return DirnBehaviourPair{elevator.D_Up, elevator.EB_Moving}
		} else if RequestsBelow(e) {
			// If no requests at the current floor, but requests below, move down
			return DirnBehaviourPair{elevator.D_Down, elevator.EB_Moving}
		}
	}
	// Default: stop the elevator if no requests
	return DirnBehaviourPair{elevator.D_Stop, elevator.EB_Idle}
}

// Check if the elevator should stop at the current floor based on requests
func RequestsShouldStop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevator.D_Down:
		// If moving down, stop if there are requests or no further requests below
		return e.Requests[e.Floor][elevator.B_HallDown] ||
			e.Requests[e.Floor][elevator.B_Cab] ||
			!RequestsBelow(e)

	case elevator.D_Up:
		// If moving up, stop if there are requests or no further requests above
		return e.Requests[e.Floor][elevator.B_HallUp] ||
			e.Requests[e.Floor][elevator.B_Cab] ||
			!RequestsAbove(e)

	case elevator.D_Stop:
		// If stopped, always return true
		fallthrough
	default:
		return true
	}
}

// Check if the current request should be cleared immediately
func RequestsShouldClearImmediately(e elevator.Elevator, btnFloor int, btnType elevator.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_All:
		// If clearing all requests, clear immediately if on the requested floor
		return e.Floor == btnFloor
	case elevator.CV_InDirn:
		// If clearing in direction, only clear if on requested floor and in the correct direction
		return e.Floor == btnFloor &&
			((e.Dirn == elevator.D_Up && btnType == elevator.B_HallUp) ||
				(e.Dirn == elevator.D_Down && btnType == elevator.B_HallDown) ||
				e.Dirn == elevator.D_Stop ||
				btnType == elevator.B_Cab)
	default:
		// No immediate clearing for other configurations
		return false
	}
}

// Clear requests at the current floor based on the configured clearing behavior
func RequestsClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_All:
		// Clear all requests at the current floor
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case elevator.CV_InDirn:
		// Clear requests at the current floor but only for the cab or direction-specific buttons
		e.Requests[e.Floor][elevator.B_Cab] = false
		switch e.Dirn {
		case elevator.D_Up:
			// If moving up, clear down request if no requests above
			if !RequestsAbove(e) && e.Requests[e.Floor][elevator.B_HallUp] == false {
				e.Requests[e.Floor][elevator.B_HallDown] = false
			}
			e.Requests[e.Floor][elevator.B_HallUp] = false
		case elevator.D_Down:
			// If moving down, clear up request if no requests below
			if !RequestsBelow(e) && e.Requests[e.Floor][elevator.B_HallDown] == false {
				e.Requests[e.Floor][elevator.B_HallUp] = false
			}
			e.Requests[e.Floor][elevator.B_HallDown] = false
		case elevator.D_Stop:
		default:
			// Clear both up and down requests for all directions
			e.Requests[e.Floor][elevator.B_HallUp] = false
			e.Requests[e.Floor][elevator.B_HallDown] = false
		}
	}
	return e
}
