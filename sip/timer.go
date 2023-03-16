package sip

import "time"

type Timer struct {
	Timer         *time.Timer
	Duration      int
	TimeoutCnt    int
	MaxTimeoutCnt int
}

func NewTimer(duration, maxTimeoutCnt int) *Timer {
	t := &Timer{
		Timer:         time.NewTimer(time.Duration(duration) * time.Second),
		Duration:      duration,
		TimeoutCnt:    0,
		MaxTimeoutCnt: maxTimeoutCnt,
	}

	return t
}

func (t *Timer) Reset(nSecond int) {
	t.Timer.Reset(time.Duration(nSecond) * time.Second)
	t.TimeoutCnt = 0
	t.Duration = nSecond
}
