package cache

import (
	"context"
	"testing"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCachedTranslationLoader_BulkLoad(t *testing.T) {
	ttl := 1 * time.Minute

	ctx := context.Background()
	entityIDs := []string{"id-1"}
	locales := []string{"en", "th"}

	// Expectation: Key is now just the entity ID
	expectedKey := "id-1"

	expectedTranslationsEN := []domain.Translation{
		{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"},
	}
	expectedTranslationsTH := []domain.Translation{
		{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"},
	}

	t.Run("Cache Miss", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		// 1. Try to get from driver — nil map = miss
		mockDriver.On("Get", ctx, expectedKey).Return(nil, nil).Once()

		// 2. Fallback to underlying
		underlyingRes := map[string][]domain.Translation{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		mockUnderlying.On("BulkLoad", ctx, entityIDs, locales).Return(underlyingRes, nil).Once()

		// 3. Store in driver as locale → Translations (fieldName → Translation)
		expectedCacheMap := map[string]domain.Translations{
			"en": {"name": {EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {"name": {EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}
		mockDriver.On("Set", ctx, expectedKey, expectedCacheMap, ttl).Return(nil).Once()

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
			"en": {"name": {EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {"name": {EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}

		// 1. Get from driver — non-nil map = hit
		mockDriver.On("Get", ctx, expectedKey).Return(cachedMap, nil).Once()

		res, err := cachedLoader.BulkLoad(ctx, entityIDs, locales)
		assert.NoError(t, err)

		expectedRes := map[string][]domain.Translation{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		assert.Equal(t, expectedRes, res)

		// Underlying should NOT be called
		mockUnderlying.AssertNotCalled(t, "BulkLoad", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Partial Cache Hit (Missing Locale)", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		// Cache only has "en"
		cachedMap := map[string]domain.Translations{
			"en": {"name": {EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
		}

		// 1. Get from driver (hit, but "th" is missing)
		mockDriver.On("Get", ctx, expectedKey).Return(cachedMap, nil).Once()

		// 2. Fallback to underlying for ALL requested locales (as per spec, it overwrites)
		underlyingRes := map[string][]domain.Translation{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		mockUnderlying.On("BulkLoad", ctx, entityIDs, locales).Return(underlyingRes, nil).Once()

		// 3. Store the full updated map in driver
		expectedCacheMap := map[string]domain.Translations{
			"en": {"name": {EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {"name": {EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}
		mockDriver.On("Set", ctx, expectedKey, expectedCacheMap, ttl).Return(nil).Once()

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

	// Test Invalidate: Should call Delete for the entityID EXACTLY ONCE
	mockDriver.On("Delete", ctx, entityID).Return(nil).Once()

	cachedLoader.Invalidate(entityID)

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
