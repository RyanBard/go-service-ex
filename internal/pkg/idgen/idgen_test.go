package idgen

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenID(t *testing.T) {
	s := New()
	actual := s.GenID()
	assert.NotEqual(t, "", actual)
	assert.Equal(t, 36, len(actual))
}
