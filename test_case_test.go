package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCase(t *testing.T) {
	c := LoadCase(t, "./tests/test-case/", nil)
	require.Len(t, c.Steps, 3)
}
