package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Compare(t *testing.T) {
	assert.Negative(t, Compare(ID(5), ID(6)))
	assert.Negative(t, Compare(nil, ID(5)))
	assert.Positive(t, Compare(ID(5), nil))
	assert.Zero(t, Compare(nil, nil))
	assert.Zero(t, Compare(ID(5), ID(5)))
}
