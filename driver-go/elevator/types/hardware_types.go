/*
Package types defines additional constants and data structures for controlling elevator motor directions
and handling button events.

Types:
- MotorDirection: Enum representing possible motor movements (Up, Down, Stop).
- ButtonType: Enum defining the types of elevator buttons (Up, Down, Cab).
- ButtonEvent: Struct representing an event when a button is pressed, including floor number, button type,
  assigned elevator, and whether the request has been completed.

These definitions support communication and decision-making in the elevator control system.
*/
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
