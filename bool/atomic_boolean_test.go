package atomicbool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean(t *testing.T) {
	atomicBoolean := NewBool(true)
	assert.True(t, atomicBoolean.IsSet())
	atomicBoolean.Unset()
	assert.False(t, atomicBoolean.IsSet())
	atomicBoolean.Set()
	assert.True(t, atomicBoolean.IsSet())
}
