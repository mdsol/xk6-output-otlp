package otlp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IDs(t *testing.T) {
	err := os.RemoveAll(varPrefix)
	require.NoError(t, err)

	for i := 1; i < 1257; i++ {
		ids, err := newIdentities()
		assert.NoError(t, err)
		val := i % 256
		assert.Equal(t, byte(ids.runID), byte(val))
	}

	err = os.RemoveAll(varPrefix)
	assert.NoError(t, err)
}
