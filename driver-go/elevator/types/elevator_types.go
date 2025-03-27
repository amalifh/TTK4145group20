/*
Package types defines the core constants and data structures used to represent elevator states,
directions, behaviors, and request handling in the elevator system.

Constants:
- N_FLOORS: The number of floors in the building.
- N_ELEVATORS: The number of elevators in the system.
- N_BUTTONS: The number of buttons per floor (Up, Down, and Cab).

Types:
  - ElevDirection: Enum representing elevator movement directions (Up, Down, Stop).
  - ElevBehaviour: Enum representing different operational states of the elevator (Idle, Moving, DoorOpen, Undefined).
  - ClearRequestVariant: Enum specifying how requests should be cleared when the elevator stops.
  - ElevInfo: Struct containing the state of an elevator, including its behavior, direction, current floor,
    request queue, and request clearing policy.

These definitions form the foundation for controlling and managing the elevator system.
*/
package types

const (
	N_FLOORS    = 4
	N_ELEVATORS = 3
	N_BUTTONS   = 3
)

type ElevDirection int

const (
	ED_Down ElevDirection = iota - 1
	ED_Stop
	ED_Up
)

type ElevBehaviour int

const (
	EB_Undefined ElevBehaviour = iota - 1
	EB_Idle
	EB_Moving
	EB_DoorOpen
)

type ClearRequestVariant int

type ElevInfo struct {
	State         ElevBehaviour
	Dir           ElevDirection
	Floor         int
	RequestsQueue [N_FLOORS][N_BUTTONS]bool
}
