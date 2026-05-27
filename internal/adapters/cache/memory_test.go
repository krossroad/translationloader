package cache

import (
	"context"
	"testing"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCachedTranslationLoader_BulkLoad(t *testing.T) {
	ttl := 1 * time.Minute

	ctx := context.Background()
	entityIDs := []string{"id-1"}
	locales := []string{"en", "th"}

	expectedKey := "id-1"

	t.Run("Cache Miss", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		underlyingRes := map[string]domain.Translations{
			"id-1": {
				{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"},
				{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"},
			},
		}

		// What the driver will return after the loader populates the cache.
		expectedCached := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}

		mockUnderlying.On("BulkLoad", mock.Anything, []string{"id-1"}, locales).Return(underlyingRes, nil).Once()

		// On a miss the driver calls the provided loader, then returns the cached value.
		mockDriver.On("Load", mock.Anything, expectedKey, ttl,
			mock.AnythingOfType("func(context.Context) (map[string]domain.Translations, error)")).
			Run(func(args mock.Arguments) {
				fn := args.Get(3).(func(context.Context) (map[string]domain.Translations, error))
				_, _ = fn(ctx) // side-effect: invokes the underlying BulkLoad
			}).
			Return(expectedCached, nil).Once()

		res, err := cachedLoader.BulkLoad(ctx, entityIDs, locales)
		assert.NoError(t, err)
		assert.Equal(t, underlyingRes, res)
		mockDriver.AssertExpectations(t)
		mockUnderlying.AssertExpectations(t)
	})

	t.Run("Full Cache Hit", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		cachedMap := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}

		// On a full hit the driver returns the cached value without invoking the loader.
		mockDriver.On("Load", mock.Anything, expectedKey, ttl,
			mock.AnythingOfType("func(context.Context) (map[string]domain.Translations, error)")).
			Return(cachedMap, nil).Once()

		res, err := cachedLoader.BulkLoad(ctx, entityIDs, locales)
		assert.NoError(t, err)

		expectedRes := map[string]domain.Translations{
			"id-1": {
				{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"},
				{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"},
			},
		}
		assert.Equal(t, expectedRes, res)
		mockUnderlying.AssertNotCalled(t, "BulkLoad", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Partial Cache Hit (Missing Locale)", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		underlyingRes := map[string]domain.Translations{
			"id-1": {
				{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"},
				{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"},
			},
		}
		expectedCached := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}

		// First Load: driver returns a partial cache hit (only "en", missing "th").
		partialCached := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
		}
		mockDriver.On("Load", mock.Anything, expectedKey, ttl,
			mock.AnythingOfType("func(context.Context) (map[string]domain.Translations, error)")).
			Return(partialCached, nil).Once()

		mockUnderlying.On("BulkLoad", mock.Anything, []string{"id-1"}, locales).Return(underlyingRes, nil).Once()

		// Partial detected: Delete evicts the stale entry, then Load fetches fresh data via the loader callback.
		mockDriver.On("Delete", mock.Anything, expectedKey).Return(nil).Once()
		mockDriver.On("Load", mock.Anything, expectedKey, ttl,
			mock.AnythingOfType("func(context.Context) (map[string]domain.Translations, error)")).
			Run(func(args mock.Arguments) {
				fn := args.Get(3).(func(context.Context) (map[string]domain.Translations, error))
				_, _ = fn(ctx) // side-effect: invokes the underlying BulkLoad
			}).
			Return(expectedCached, nil).Once()

		res, err := cachedLoader.BulkLoad(ctx, entityIDs, locales)
		assert.NoError(t, err)
		assert.Equal(t, underlyingRes, res)
		mockDriver.AssertExpectations(t)
		mockUnderlying.AssertExpectations(t)
	})
}

func TestCachedTranslationLoader_Invalidate_O1(t *testing.T) {
	mockUnderlying := mocks.NewTranslationLoader(t)
	mockDriver := mocks.NewCacheDriver(t)
	ttl := 1 * time.Minute
	cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

	ctx := context.Background()
	entityID := "PROD-1"

	mockDriver.On("Delete", ctx, entityID).Return(nil).Once()

	err := cachedLoader.Invalidate(entityID)
	assert.NoError(t, err)

	mockDriver.AssertExpectations(t)
}

func TestCachedTranslationLoader_Invalidate_ReturnsError(t *testing.T) {
	// Bug 2: Invalidate discards the Delete error with `_ = ...`, so callers cannot
	// detect a failed invalidation. The desired signature is Invalidate(entityID string) error.
	mockUnderlying := mocks.NewTranslationLoader(t)
	mockDriver := mocks.NewCacheDriver(t)
	ttl := 1 * time.Minute
	cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

	entityID := "PROD-42"
	deleteErr := assert.AnError

	mockDriver.On("Delete", context.Background(), entityID).Return(deleteErr).Once()

	// Invalidate must propagate the Delete error to the caller.
	err := cachedLoader.Invalidate(entityID)
	assert.ErrorIs(t, err, deleteErr, "Invalidate must return the error from the cache driver's Delete call")

	mockDriver.AssertExpectations(t)
}
