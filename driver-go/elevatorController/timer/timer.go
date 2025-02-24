package timer

import "time"

var (
	timerEndTime time.Time // Time when the timer will end
	timerActive  bool      // Flag to indicate whether the timer is active
)

// TimerStart initializes the timer with the given duration in seconds and sets it as active.
func TimerStart(duration float64) {
	// Set the end time by adding the duration to the current time
	timerEndTime = time.Now().Add(time.Duration(duration * float64(time.Second)))
	// Mark the timer as active
	timerActive = true
}

// TimerStop deactivates the timer, effectively stopping it.
func TimerStop() {
	// Set the timer as inactive
	timerActive = false
}

// TimerTimedOut checks if the timer has expired
func TimerTimedOut() bool {
	// Return true if the timer is active and the current time has passed the end time
	return timerActive && time.Now().After(timerEndTime)
}
