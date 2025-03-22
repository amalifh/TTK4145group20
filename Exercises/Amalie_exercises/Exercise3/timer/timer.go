package timer

import "time"

var (
	timerEndTime time.Time
	timerActive  bool
)

func TimerStart(duration float64) {
	timerEndTime = time.Now().Add(time.Duration(duration * float64(time.Second)))
	timerActive = true
}

func TimerStop() bool {
	return timerActive == false
}

// checking if timer has expired
func TimerTimedOut() bool {
	return timerActive && (time.Now().After(timerEndTime))
}
