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

	expectedKey := "id-1"

	expectedTranslationsEN := domain.Translations{
		{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"},
	}
	expectedTranslationsTH := domain.Translations{
		{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"},
	}

	t.Run("Cache Miss", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		// nil map = miss
		mockDriver.On("Get", ctx, expectedKey).Return(nil, nil).Once()

		underlyingRes := map[string]domain.Translations{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		mockUnderlying.On("BulkLoad", ctx, entityIDs, locales).Return(underlyingRes, nil).Once()

		// Stored in cache grouped by locale
		expectedCacheMap := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
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
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
		}

		// non-nil map = hit
		mockDriver.On("Get", ctx, expectedKey).Return(cachedMap, nil).Once()

		res, err := cachedLoader.BulkLoad(ctx, entityIDs, locales)
		assert.NoError(t, err)

		expectedRes := map[string]domain.Translations{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		assert.Equal(t, expectedRes, res)

		mockUnderlying.AssertNotCalled(t, "BulkLoad", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Partial Cache Hit (Missing Locale)", func(t *testing.T) {
		mockUnderlying := mocks.NewTranslationLoader(t)
		mockDriver := mocks.NewCacheDriver(t)
		cachedLoader := NewCachedTranslationLoader(mockUnderlying, mockDriver, ttl)

		// Cache only has "en"
		cachedMap := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
		}

		mockDriver.On("Get", ctx, expectedKey).Return(cachedMap, nil).Once()

		underlyingRes := map[string]domain.Translations{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		mockUnderlying.On("BulkLoad", ctx, entityIDs, locales).Return(underlyingRes, nil).Once()

		expectedCacheMap := map[string]domain.Translations{
			"en": {{EntityID: "id-1", Locale: "en", FieldName: "name", FieldValue: "Name EN"}},
			"th": {{EntityID: "id-1", Locale: "th", FieldName: "name", FieldValue: "Name TH"}},
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
