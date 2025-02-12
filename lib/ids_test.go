package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeRandomID(t *testing.T) {
	r := FariHazardanID("abc", 3)

	assert.Len(t, r, 7)
}
