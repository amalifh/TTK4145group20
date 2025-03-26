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

func copyAckList(
	msg types.NetworkMessage, 
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