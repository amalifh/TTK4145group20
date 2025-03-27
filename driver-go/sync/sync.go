/*
Package sync manages synchronization and communication between distributed elevators in a multi-elevator control system.

This package provides:
- `SyncChannels`: A struct that defines channels for inter-process communication, handling updates to request assignments, elevator states, network messages, and peer connections.
- `Synchronise`: A function that ensures consistent state synchronization across elevators by exchanging messages, handling peer updates, and managing request acknowledgments.

The synchronization mechanism ensures that elevator assignments remain consistent across the network, accounts for elevator failures, and facilitates reassignment of lost requests to maintain system reliability.

Credits: https://github.com/perkjelsvik/TTK4145-sanntid
*/
package sync

import (
	"time"

	"Driver-go/elevator/types"
	"Driver-go/network/peers"
)

type SyncChannels struct {
	UpdateReqAssigner chan [types.N_ELEVATORS]types.ElevInfo
	UpdateSync        chan types.ElevInfo
	RequestsUpdate    chan types.ButtonEvent
	AliveElevators    chan [types.N_ELEVATORS]bool
	IncomingMsg       chan types.NetworkMessage
	OutgoingMsg       chan types.NetworkMessage
	PeerUpdate        chan peers.PeerUpdate
	PeerTxEnable      chan bool
}

func Synchronise(ch SyncChannels, id int) {
	var (
		registeredRequests [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList
		elevList           [types.N_ELEVATORS]types.ElevInfo
		sendMsg            types.NetworkMessage
		aliveList          [types.N_ELEVATORS]bool
		recentlyDied       [types.N_ELEVATORS]bool
		someUpdate         bool
		offline            bool
	)

	initializeSync(&elevList, &registeredRequests, &someUpdate, &offline, ch, id)

	timers := struct {
		reassign   *time.Timer
		broadcast  *time.Ticker
		singleMode *time.Ticker
		reInit     *time.Timer
	}{
		reassign:   time.NewTimer(5 * time.Second),
		broadcast:  time.NewTicker(100 * time.Millisecond),
		singleMode: time.NewTicker(100 * time.Millisecond),
		reInit:     time.NewTimer(1 * time.Second),
	}
	defer timers.reassign.Stop()
	defer timers.broadcast.Stop()
	timers.reassign.Stop()
	timers.singleMode.Stop()

	lostID := -1

	for {
		handleOfflineState(&offline, aliveList, id, timers.reInit, ch, &elevList, &someUpdate)
		handleElevatorDeath(&lostID, &recentlyDied)

		select {
		case newElev := <-ch.UpdateSync:
			handleElevatorUpdate(newElev, &elevList, id, ch.PeerTxEnable, &someUpdate)

		case newRequest := <-ch.RequestsUpdate:
			handleRequestUpdate(newRequest, &elevList, id, &registeredRequests, &someUpdate)

		case msg := <-ch.IncomingMsg:
			processIncomingMessage(msg, id, &elevList, &registeredRequests, &someUpdate, aliveList, ch.UpdateReqAssigner)

		case <-timers.singleMode.C:
			handleSingleModeOperations(&registeredRequests, &elevList, id, &someUpdate, ch.UpdateReqAssigner)

		case <-timers.broadcast.C:
			broadcastState(&sendMsg, registeredRequests, elevList, id, offline, ch.OutgoingMsg)

		case peerUpdate := <-ch.PeerUpdate:
			lostID = handlePeerUpdate(peerUpdate, &aliveList, &recentlyDied, &offline,
				&timers.singleMode, elevList, ch.AliveElevators)

		case <-timers.reassign.C:
			handleRequestReassignment(&elevList, &recentlyDied, id, ch.UpdateReqAssigner)
		}
	}
}
