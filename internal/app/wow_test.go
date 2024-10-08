package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	t.Run("GetNextWoW", func(t *testing.T) {
		t.Run("should get next random phrase", func(t *testing.T) {
			query := NewWowQuery()

			phrase1, err := query.GetNextWoW(context.Background())
			require.NoError(t, err)

			phrase2, err := query.GetNextWoW(context.Background())
			require.NoError(t, err)
			assert.NotEqual(t, phrase1, phrase2)
		})
	})
}
