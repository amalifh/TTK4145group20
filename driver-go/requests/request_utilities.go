/*request_utilities*/
package requests

import . "driver-go/elevator/types"

// DuplicateRequest checks if the request is already in the queue
func duplicateRequest(request ButtonEvent, elevList [N_ELEVATORS]ElevInfo, id int) bool {
	if request.Btn == types.BT_Cab && elevList[id].RequestsQueue[request.Floor][types.BT_Cab] {
		return true
	}
	for elevator := 0; elevator < N_ELEVATORS; elevator++ {
		if elevList[id].RequestsQueue[request.Floor][request.Btn] {
			return true
		}
	}
	return false
}

// CalcChosenElevator calculates the best elevator to handle the request
// based on the current state of the elevators
func calcChosenElevator(request ButtonEvent, elevList [N_ELEVATORS]ElevInfo, id int, aliveList [N_ELEVATORS]bool) int {
	if request.Btn == types.BT_Cab {
		return id
	}

	minCost := (N_BUTTONS * N_FLOORS) * N_ELEVATORS
	bestElevator := id
	for elevator := 0; elevator < N_ELEVATORS; elevator++ {
		if !aliveList[elevator] {
			// Disregarding offline elevators
			continue
		}
		cost := request.Floor - elevList[elevator].Floor

		// If the elevator is idle and at the same floor as the request
		// it should be chosen
		if cost == 0 && elevList[elevator].State != types.EB_Moving {
			bestElevator = elevator
			return bestElevator
		}

		// If the elevator is moving in the same direction as the request
		// the cost is reduced
		if cost < 0 {
			cost = -cost
			if elevList[elevator].Dir == types.ED_Up {
				cost += 3
			}
			// If the elevator is moving in the opposite direction as the request
			// the cost is increased
		} else if cost > 0 {
			if elevList[elevator].Dir == types.ED_Down {
				cost += 3
			}
		}

		// If the elevator is moving and the request is at the same floor
		// the cost is increased
		if cost == 0 && elevList[elevator].State == types.EB_Moving {
			cost += 4
		}

		// If the elevator's doors are open the cost is increased
		if elevList[elevator].State == types.EB_DoorOpen {
			cost++
		}

		// If the cost of the current elevator is lower than the current best
		// elevator, the current elevator is chosen
		if cost < minCost {
			minCost = cost
			bestElevator = elevator
		}
	}
	return bestElevator
}
