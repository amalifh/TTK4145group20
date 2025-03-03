package types

type RequestState int
const (
	COMPLETED RequestState = iota
	NEW
	ASSIGNED 
)

type Request struct{
	State RequestState
	Count int
	AwareList []string
}

type NetworkMessage struct{
	SID string	// Sender ID
	Available bool
	Behaviour ElevBehaviour
	Direction ElevDirection
	Floor int
	SHallRequests [N_FLOORS][N_HALL_BUTTONS]Request
	AllCabRequest map[string][N_FLOORS]Request
}