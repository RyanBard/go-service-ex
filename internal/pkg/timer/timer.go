package timer

import "time"

type timer struct{}

func New() *timer {
	return &timer{}
}

func (t *timer) Now() int64 {
	return time.Now().UnixMilli()
}
