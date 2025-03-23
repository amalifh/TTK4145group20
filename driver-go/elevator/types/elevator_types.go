package types

const (
	N_FLOORS    = 4
	N_ELEVATORS = 3
	N_BUTTONS   = 3
)

type ElevDirection int

const (
	ED_Down ElevDirection = iota - 1 // Moving down
	ED_Stop                          // Stopped (idle state)
	ED_Up                            // Moving up
)

type ElevBehaviour int

const (
	EB_Undefined ElevBehaviour = iota - 1
	EB_Idle
	EB_Moving
	EB_DoorOpen
)

type ClearRequestVariant int

const (
	CV_All    ClearRequestVariant = iota // Everyone enters the elevator, even if going in the "wrong" direction
	CV_InDirn                            // Only passengers traveling in the current direction enter
)

type ElevInfo struct {
	State         ElevBehaviour
	Dir           ElevDirection
	Floor         int
	RequestsQueue [N_FLOORS][N_BUTTONS]bool
	CV            ClearRequestVariant
}
