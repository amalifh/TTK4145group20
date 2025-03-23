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
