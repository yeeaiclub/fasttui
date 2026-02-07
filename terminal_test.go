package fasttui

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetSize(t *testing.T) {
	term, err := NewProcessTerminal()
	require.NoError(t, err)
	width, height := term.GetSize()
	assert.Greater(t, width, 0)
	assert.Greater(t, height, 0)
}
