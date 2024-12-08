package idgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenID(t *testing.T) {
	s := New()
	actual := s.GenID()
	assert.NotEqual(t, "", actual)
	assert.Equal(t, 36, len(actual))
}
