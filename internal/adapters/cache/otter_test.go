package cache

import (
	"context"
	"sync/atomic"
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

// TestOtterDriver_NoStaleWriteAfterDelete verifies that a Delete issued while a
// loader is in-flight does not result in a stale value being written to the cache.
//
// Sequence:
//  1. Call Load("k3") — miss, loader starts.
//  2. Loader blocks on a gate channel.
//  3. While loader is blocked, call Delete("k3").
//  4. Release gate — loader returns stale value.
//  5. Assert next Load("k3") invokes the loader again (no stale write landed).
func TestOtterDriver_NoStaleWriteAfterDelete(t *testing.T) {
	ctx := context.Background()
	ttl := 1 * time.Minute
	capacity := 10

	driver, err := NewOtterDriver(capacity, ttl)
	require.NoError(t, err)

	gate := make(chan struct{})
	loaderDone := make(chan struct{})

	var loaderCallCount int64

	go func() {
		_, _ = driver.Load(ctx, "k3", ttl, func(lctx context.Context) (map[string]domain.Translations, error) {
			atomic.AddInt64(&loaderCallCount, 1)
			<-gate
			return map[string]domain.Translations{
				"en": {{EntityID: "E3", Locale: "en", FieldName: "label", FieldValue: "stale"}},
			}, nil
		})
		close(loaderDone)
	}()

	// Give goroutine time to enter the loader and block.
	time.Sleep(20 * time.Millisecond)

	// Delete fires while loader is still blocked.
	require.NoError(t, driver.Delete(ctx, "k3"))

	// Release the loader — it returns stale data and must NOT write to cache.
	close(gate)
	<-loaderDone

	// A second Load must invoke the loader again — the cache must be empty.
	var secondLoaderCalled int64
	got, err := driver.Load(ctx, "k3", ttl, func(lctx context.Context) (map[string]domain.Translations, error) {
		atomic.AddInt64(&secondLoaderCalled, 1)
		return map[string]domain.Translations{
			"en": {{EntityID: "E3", Locale: "en", FieldName: "label", FieldValue: "fresh"}},
		}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), atomic.LoadInt64(&secondLoaderCalled),
		"loader must be called after Delete — stale write must have been prevented")
	assert.Equal(t, "fresh", got["en"][0].FieldValue,
		"returned value must be from the fresh load, not the stale in-flight one")
}
