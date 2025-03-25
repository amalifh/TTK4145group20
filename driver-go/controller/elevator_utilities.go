package controller

import (
	"Driver-go/elevator/types"
)

func shouldStop(elevator types.ElevInfo) bool {
	switch elevator.Dir {
	case types.ED_Up:
		return elevator.RequestsQueue[elevator.Floor][types.BT_Up] ||
			elevator.RequestsQueue[elevator.Floor][types.BT_Cab] ||
			!isRequestAbove(elevator)
	case types.ED_Down:
		return elevator.RequestsQueue[elevator.Floor][types.BT_Down] ||
			elevator.RequestsQueue[elevator.Floor][types.BT_Cab] ||
			!isRequestBelow(elevator)
	case types.ED_Stop:
	default:
	}
	return false
}

func chooseDirection(elevator types.ElevInfo) types.ElevDirection {
	switch elevator.Dir {
	case types.ED_Stop:
		if isRequestAbove(elevator) {
			return types.ED_Up
		} else if isRequestBelow(elevator) {
			return types.ED_Down
		} else {
			return types.ED_Stop
		}

	case types.ED_Up:
		if isRequestAbove(elevator) {
			return types.ED_Up
		} else if isRequestBelow(elevator) {
			return types.ED_Down
		} else {
			return types.ED_Stop
		}

	case types.ED_Down:
		if isRequestBelow(elevator) {
			return types.ED_Down
		} else if isRequestAbove(elevator) {
			return types.ED_Up
		} else {
			return types.ED_Stop
		}
	}
	return types.ED_Stop
}

func DirectionConverter(dir types.ElevDirection) types.MotorDirection {
	switch dir {
	case types.ED_Up:
		return types.MD_Up
	case types.ED_Down:
		return types.MD_Down
	case types.ED_Stop:
		return types.MD_Stop
	}
	return types.MD_Stop
}

func isRequestAbove(elevator types.ElevInfo) bool {
	for floor := elevator.Floor + 1; floor < types.N_FLOORS; floor++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if elevator.RequestsQueue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func isRequestBelow(elevator types.ElevInfo) bool {
	for floor := 0; floor < elevator.Floor; floor++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if elevator.RequestsQueue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func clearRequests(elevator types.ElevInfo, floor int) types.ElevInfo {
	switch elevator.CV {
	case types.CV_All:
		// Clear all buttons at the current floor
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			elevator.RequestsQueue[floor][btn] = false
		}
	case types.CV_InDirn:
		// Always clear Cab request
		elevator.RequestsQueue[floor][types.BT_Cab] = false

		switch elevator.Dir {
		case types.ED_Up:
			// Check if no requests above and current HallUp is not active
			if !isRequestAbove(elevator) && !elevator.RequestsQueue[floor][types.BT_Up] {
				elevator.RequestsQueue[floor][types.BT_Down] = false
			}
			// Clear HallUp
			elevator.RequestsQueue[floor][types.BT_Up] = false

		case types.ED_Down:
			// Check if no requests below and current HallDown is not active
			if !isRequestBelow(elevator) && !elevator.RequestsQueue[floor][types.BT_Down] {
				elevator.RequestsQueue[floor][types.BT_Up] = false
			}
			// Clear HallDown
			elevator.RequestsQueue[floor][types.BT_Down] = false

		case types.ED_Stop:
			fallthrough
		default:
			// Clear both HallUp and HallDown
			elevator.RequestsQueue[floor][types.BT_Up] = false
			elevator.RequestsQueue[floor][types.BT_Down] = false
		}
	}
	return elevator
}
