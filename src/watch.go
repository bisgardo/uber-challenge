package uber_challenge

import "time"

type StopWatch struct {
	InitTime  time.Time
	StartTime time.Time
}

func NewStopWatch() (sw StopWatch) {
	sw.InitTime = time.Now()
	sw.StartTime = sw.InitTime
	return
}

const MS_PER_NS = 1000000

func (sw *StopWatch) TotalElapsedTimeMillis() int64 {
	return (time.Now().UnixNano() - sw.InitTime.UnixNano()) / MS_PER_NS
}

func (sw *StopWatch) ElapsedTimeMillis(reset bool) int64 {
	now := time.Now()
	ms := (now.UnixNano() - sw.StartTime.UnixNano()) / MS_PER_NS
	if reset {
		sw.StartTime = now
	}
	return ms
}

