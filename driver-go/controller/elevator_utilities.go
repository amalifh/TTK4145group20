package controller

import (
	"Driver-go/elevator/types"
	"sort"
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


/*
https://github.com/TTK4145/Project-resources/tree/master/elev_algo
func clearRequests(elevator types.ElevInfo, startingFloor int) types.ElevInfo {

}
*/