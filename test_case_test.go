package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCase(t *testing.T) {
	c := LoadCase(t, "./tests")
	require.Len(t, c.Steps, 2)
}
