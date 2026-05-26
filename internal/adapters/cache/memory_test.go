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

		// 1. Try to get from driver
		mockDriver.On("Get", ctx, expectedKey).Return(nil, false, nil).Once()
		
		// 2. Fallback to underlying
		underlyingRes := map[string][]domain.Translation{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		mockUnderlying.On("BulkLoad", ctx, entityIDs, locales).Return(underlyingRes, nil).Once()
		
		// 3. Store in driver as a map keyed by locale
		expectedCacheMap := map[string][]domain.Translation{
			"en": expectedTranslationsEN,
			"th": expectedTranslationsTH,
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

		cachedMap := map[string][]domain.Translation{
			"en": expectedTranslationsEN,
			"th": expectedTranslationsTH,
		}

		// 1. Get from driver
		mockDriver.On("Get", ctx, expectedKey).Return(cachedMap, true, nil).Once()

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
		cachedMap := map[string][]domain.Translation{
			"en": expectedTranslationsEN,
		}

		// 1. Get from driver (hit, but "th" is missing)
		mockDriver.On("Get", ctx, expectedKey).Return(cachedMap, true, nil).Once()

		// 2. Fallback to underlying for ALL requested locales (as per spec, it overwrites)
		underlyingRes := map[string][]domain.Translation{
			"id-1": append(expectedTranslationsEN, expectedTranslationsTH...),
		}
		mockUnderlying.On("BulkLoad", ctx, entityIDs, locales).Return(underlyingRes, nil).Once()

		// 3. Store the full updated map in driver
		expectedCacheMap := map[string][]domain.Translation{
			"en": expectedTranslationsEN,
			"th": expectedTranslationsTH,
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
