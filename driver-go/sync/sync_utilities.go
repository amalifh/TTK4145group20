/*
Package sync provides functions for managing synchronization and acknowledgment of elevator requests in a distributed elevator system.

This package includes:
- `copyAckList`: Copies acknowledgment data from a network message to the local registered orders, ensuring consistency across elevators.
- `checkAllAckStatus`: Verifies whether all active elevators have a specific acknowledgment status for a given request.

These functions help maintain synchronization between elevators by tracking implicit acknowledgments and ensuring correct request assignments.

Credits: https://github.com/perkjelsvik/TTK4145-sanntid
*/
package sync

import "Driver-go/elevator/types"

func copyAckList(msg types.NetworkMessage,
	registeredOrders [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	elevator, floor, id int,
	btn types.ButtonType) [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList {

	registeredOrders[floor][btn].ImplicitAcks[id] = msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ImplicitAcks[elevator] = msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ChosenElevator = msg.RegisteredRequests[floor][btn].ChosenElevator
	return registeredOrders
}

func checkAllAckStatus(requestsList [types.N_ELEVATORS]bool, ImplicitAcks [types.N_ELEVATORS]types.Acknowledge, status types.Acknowledge) bool {
	for elev := 0; elev < types.N_ELEVATORS; elev++ {
		if !requestsList[elev] {
			continue
		}
		if ImplicitAcks[elev] != status {
			return false
		}
	}
	return true
}

func shouldProcessMessage(msg types.NetworkMessage, id int, aliveList [types.N_ELEVATORS]bool) bool {
	return msg.ID != id && aliveList[msg.ID] && aliveList[id]
}

func updateElevatorState(msg types.NetworkMessage, id int, elevList [types.N_ELEVATORS]types.ElevInfo) ([types.N_ELEVATORS]types.ElevInfo, bool) {
	if msg.Elevator != elevList {

		tmpElevator := elevList[id]
		newElevList := msg.Elevator
		newElevList[id] = tmpElevator
		return newElevList, true
	}
	return elevList, false
}

func processAcksForElevator(
	msg types.NetworkMessage,
	elevator, id int,
	registeredRequests [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	elevList [types.N_ELEVATORS]types.ElevInfo,
	aliveList [types.N_ELEVATORS]bool,
) ([types.N_FLOORS][types.N_BUTTONS - 1]types.AckList, [types.N_ELEVATORS]types.ElevInfo, bool) {
	updateOccurred := false

	for floor := 0; floor < types.N_FLOORS; floor++ {
		for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
			remoteAck := msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator]
			localAck := registeredRequests[floor][btn].ImplicitAcks
			switch remoteAck {
			case types.NOTACK:
				if localAck[id] == types.COMPLETED {
					registeredRequests = copyAckList(msg, registeredRequests, elevator, floor, id, btn)
				} else if localAck[elevator] != types.NOTACK {
					registeredRequests[floor][btn].ImplicitAcks[elevator] = types.NOTACK
				}
			case types.ACK:
				if localAck[id] == types.NOTACK {
					registeredRequests = copyAckList(msg, registeredRequests, elevator, floor, id, btn)
				} else if localAck[elevator] != types.ACK {
					registeredRequests[floor][btn].ImplicitAcks[elevator] = types.ACK
				}
				if checkAllAckStatus(aliveList, registeredRequests[floor][btn].ImplicitAcks, types.ACK) &&
					!elevList[id].RequestsQueue[floor][btn] &&
					registeredRequests[floor][btn].ChosenElevator == id {
					elevList[id].RequestsQueue[floor][btn] = true
					updateOccurred = true
				}
			case types.COMPLETED:
				if localAck[id] == types.ACK {
					registeredRequests = copyAckList(msg, registeredRequests, elevator, floor, id, btn)
				} else if localAck[elevator] != types.COMPLETED {
					registeredRequests[floor][btn].ImplicitAcks[elevator] = types.COMPLETED
				}
				if checkAllAckStatus(aliveList, registeredRequests[floor][btn].ImplicitAcks, types.COMPLETED) {
					registeredRequests[floor][btn].ImplicitAcks[id] = types.NOTACK
					if registeredRequests[floor][btn].ChosenElevator == id {
						elevList[id].RequestsQueue[floor][btn] = false
						updateOccurred = true
					}
				}
			}
		}
	}
	return registeredRequests, elevList, updateOccurred
}
