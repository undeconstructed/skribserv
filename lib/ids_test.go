package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeRandomID(t *testing.T) {
	r := MakeRandomID("abc", 3)

	assert.Len(t, r, 7)
}
