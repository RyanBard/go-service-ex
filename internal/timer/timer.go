package timer

import "time"

type timer struct{}

func New() *timer {
	return &timer{}
}

func (t *timer) Now() time.Time {
	return time.Now()
}
