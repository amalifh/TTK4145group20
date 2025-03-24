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