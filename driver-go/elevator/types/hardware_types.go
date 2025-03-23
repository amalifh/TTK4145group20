package types

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type ButtonType int

const (
	BT_Up ButtonType = iota
	BT_Down
	BT_Cab
)

type ButtonEvent struct {
	Floor          int
	Btn            ButtonType
	ChosenElevator int
	Done           bool
}
