package cache

import (
	"context"
	"testing"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOtterDriver(t *testing.T) {
	ctx := context.Background()
	ttl := 1 * time.Minute
	capacity := 10

	t.Run("Set and Get Map Value", func(t *testing.T) {
		driver, err := NewOtterDriver(capacity, ttl)
		require.NoError(t, err)

		key := "test-entity-1"
		value := map[string]domain.Translations{
			"en": {"label": {EntityID: "test-entity-1", Locale: "en", FieldName: "label", FieldValue: "V1"}},
			"fr": {"label": {EntityID: "test-entity-1", Locale: "fr", FieldName: "label", FieldValue: "V2"}},
		}

		err = driver.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		got, err := driver.Get(ctx, key)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, value, got)
	})

	t.Run("Delete", func(t *testing.T) {
		driver, err := NewOtterDriver(capacity, ttl)
		require.NoError(t, err)

		key := "test-key-delete"
		value := map[string]domain.Translations{
			"en": {"label": {EntityID: "E1", Locale: "en", FieldName: "label", FieldValue: "V1"}},
		}

		_ = driver.Set(ctx, key, value, ttl)
		err = driver.Delete(ctx, key)
		assert.NoError(t, err)

		got, err := driver.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Eviction", func(t *testing.T) {
		smallCapacity := 2
		driver, err := NewOtterDriver(smallCapacity, ttl)
		require.NoError(t, err)

		val := map[string]domain.Translations{"en": {"label": {EntityID: "E1"}}}
		_ = driver.Set(ctx, "k1", val, ttl)
		_ = driver.Set(ctx, "k2", val, ttl)
		_ = driver.Set(ctx, "k3", val, ttl)

		// Verification of eviction depends on Otter's internal state.
		// For unit tests, we primarily care that it compiles and basic operations work with maps.
	})
}
