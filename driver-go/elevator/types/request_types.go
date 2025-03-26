/*
Package types provides definitions for network communication and acknowledgment mechanisms
in a multi-elevator system.

Types:
- Acknowledge: Enum representing the acknowledgment status (COMPLETED, NOTACK, ACK).
- AckList: Struct containing the chosen elevator for a request and implicit acknowledgments
  from all elevators in the system.
- NetworkMessage: Struct representing a network message containing the current state of all elevators,
  registered requests with acknowledgments, and an identifier.

These definitions facilitate communication and coordination between multiple elevators
in a distributed control system.
*/
package types

type Acknowledge int

const (
	COMPLETED Acknowledge = iota - 1
	NOTACK
	ACK
)

type AckList struct {
	ChosenElevator int
	ImplicitAcks   [N_ELEVATORS]Acknowledge
}

type NetworkMessage struct {
	Elevator           [N_ELEVATORS]ElevInfo
	RegisteredRequests [N_FLOORS][N_BUTTONS - 1]AckList
	ID                 int
}
