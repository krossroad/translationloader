package cache

import (
	"context"
	"testing"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOtterDriver_Load verifies the new Load method on the otter driver.
//
// NOTE: This test will NOT compile until ports.CacheDriver gains a Load method
// and otterDriver implements it. That compile failure is intentional — it
// documents the new interface requirement.
func TestOtterDriver_Load(t *testing.T) {
	ctx := context.Background()
	ttl := 1 * time.Minute
	capacity := 10

	t.Run("Load Miss then Hit", func(t *testing.T) {
		driver, err := NewOtterDriver(capacity, ttl)
		require.NoError(t, err)

		callCount := 0
		loader := func(lctx context.Context) (map[string]domain.Translations, error) {
			callCount++
			return map[string]domain.Translations{
				"en": {{EntityID: "E1", Locale: "en", FieldName: "label", FieldValue: "V1"}},
			}, nil
		}

		// First call: miss, loader invoked.
		got, err := driver.Load(ctx, "k1", ttl, loader)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, 1, callCount)

		// Second call: hit, loader NOT invoked again.
		got2, err := driver.Load(ctx, "k1", ttl, loader)
		assert.NoError(t, err)
		assert.Equal(t, got, got2)
		assert.Equal(t, 1, callCount, "loader must not be called on cache hit")
	})

	t.Run("Delete clears entry", func(t *testing.T) {
		driver, err := NewOtterDriver(capacity, ttl)
		require.NoError(t, err)

		loaded := map[string]domain.Translations{
			"en": {{EntityID: "E2", Locale: "en", FieldName: "label", FieldValue: "V1"}},
		}
		_, _ = driver.Load(ctx, "k2", ttl, func(lctx context.Context) (map[string]domain.Translations, error) {
			return loaded, nil
		})

		assert.NoError(t, driver.Delete(ctx, "k2"))

		callCount := 0
		_, _ = driver.Load(ctx, "k2", ttl, func(lctx context.Context) (map[string]domain.Translations, error) {
			callCount++
			return nil, nil
		})
		assert.Equal(t, 1, callCount, "loader must be called after Delete")
	})
}
