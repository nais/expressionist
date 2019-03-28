package expressionist

import (
	"github.com/nais/expressionist/pkg/expressionist/fixtures"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseAlert(t *testing.T) {

	t.Run("No output when valid", func(t *testing.T) {
		output, err := ParseAlert(fixtures.ValidConfiguration)
		assert.NoError(t, err)
		assert.Empty(t, output)
	})

	t.Run("Some output when not valid", func(t *testing.T) {
		output, err := ParseAlert(fixtures.NotValidConfiguration)
		assert.NoError(t, err)
		assert.NotEmpty(t, output)
	})
}
