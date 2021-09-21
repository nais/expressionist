package expressionist

import (
	"testing"

	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/stretchr/testify/assert"

	"github.com/nais/expressionist/pkg/expressionist/fixtures"
)

func TestParseAlert(t *testing.T) {

	t.Run("No output when valid", func(t *testing.T) {
		for _, test := range []struct {
			description string
			rules       *naisiov1.Alert
		}{
			{
				description: "empty configuration",
				rules:       fixtures.EmptySpec(),
			},
			{
				description: "valid configuration",
				rules:       fixtures.ValidConfiguration(),
			},
		} {
			t.Run(test.description, func(t *testing.T) {
				output, err := ValidateRules(test.rules)
				assert.NoError(t, err)
				assert.Empty(t, output)
			})
		}
	})

	t.Run("Some output when not valid", func(t *testing.T) {
		for _, test := range []struct {
			description string
			rules       *naisiov1.Alert
		}{
			{
				description: "invalid expr",
				rules:       fixtures.InvalidExpr(),
			},
			{
				description: "invalid action",
				rules:       fixtures.InvalidAction(),
			},
			{
				description: "invalid description",
				rules:       fixtures.InvalidDescription(),
			},
		} {
			t.Run(test.description, func(t *testing.T) {
				output, err := ValidateRules(test.rules)
				assert.NoError(t, err)
				assert.NotEmpty(t, output)
			})
		}
	})
}
