/*
Package peers provides functionality for managing peer-to-peer communication in a networked system. It allows for the transmission of peer IDs via UDP broadcast and tracks the status of peers, including detecting newly added or lost peers over time.

Key Features:
- Transmitter: Periodically broadcasts the peer's ID to a multicast address, enabling other peers to discover and track it.
- Receiver: Listens for incoming broadcast messages from peers, identifies new and lost peers, and updates the list of active peers.
- Peer Update: The package provides real-time updates on the current state of peers, including newly discovered peers and those that have timed out.

Constants:
- `interval`: Defines the period between transmission attempts (15 milliseconds).
- `timeout`: Specifies the timeout duration to consider a peer as lost (500 milliseconds).

Functions:
- Transmitter: Takes a UDP port, a peer ID, and a channel that enables transmission. Periodically sends the peer's ID to all peers.
- Receiver: Listens for incoming peer IDs on the specified port, tracks the arrival and departure of peers, and sends updates on the current peer list and any lost peers.

Usage:
- To transmit a peer ID, call the Transmitter function with the desired port, peer ID, and a channel that controls whether transmission is enabled.
- To receive peer updates, call the Receiver function with the desired port and a channel to receive `PeerUpdate` messages.

Note:
This package uses UDP broadcast to send and receive peer IDs, meaning it requires the network to support UDP broadcasts (e.g., local networks).

Credits: https://github.com/TTK4145/Network-go
*/
package peers

import (
	"Driver-go/network/conn"
	"fmt"
	"net"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}