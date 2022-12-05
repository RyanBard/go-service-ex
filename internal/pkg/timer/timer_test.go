package timer

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	now := time.Now().UnixMilli()
	s := New()
	time.Sleep(10 * time.Millisecond)
	actual1 := s.Now()
	assert.Greater(t, actual1.UnixMilli(), now)
	time.Sleep(10 * time.Millisecond)
	actual2 := s.Now()
	assert.Greater(t, actual2.UnixMilli(), actual1.UnixMilli())
}
