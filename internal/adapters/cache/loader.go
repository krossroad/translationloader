package cache

import (
	"context"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/ports"
)

type CachedTranslationLoader struct {
	underlying ports.TranslationLoader
	driver     ports.CacheDriver
	ttl        time.Duration
}

func NewCachedTranslationLoader(underlying ports.TranslationLoader, driver ports.CacheDriver, ttl time.Duration) *CachedTranslationLoader {
	return &CachedTranslationLoader{
		underlying: underlying,
		driver:     driver,
		ttl:        ttl,
	}
}

func (c *CachedTranslationLoader) BulkLoad(ctx context.Context, entityIDs []string, locales []string) (map[string][]domain.Translation, error) {
	results := make(map[string][]domain.Translation)
	var missingIDs []string

	for _, id := range entityIDs {
		cachedMap, found, err := c.driver.Get(ctx, id)
		if err != nil || !found {
			missingIDs = append(missingIDs, id)
			continue
		}

		// Verify all locales are present
		allPresent := true
		for _, locale := range locales {
			if _, ok := cachedMap[locale]; !ok {
				allPresent = false
				break
			}
		}

		if !allPresent {
			missingIDs = append(missingIDs, id)
			continue
		}

		// Combine cached translations for the requested locales
		var entityTrans []domain.Translation
		for _, locale := range locales {
			entityTrans = append(entityTrans, cachedMap[locale]...)
		}
		results[id] = entityTrans
	}

	if len(missingIDs) > 0 {
		fresh, err := c.underlying.BulkLoad(ctx, missingIDs, locales)
		if err != nil {
			return nil, err
		}

		for id, trans := range fresh {
			results[id] = trans

			// Store in cache grouped by locale
			grouped := make(map[string][]domain.Translation)
			for _, t := range trans {
				grouped[t.Locale] = append(grouped[t.Locale], t)
			}
			_ = c.driver.Set(ctx, id, grouped, c.ttl)
		}
	}

	return results, nil
}

func (c *CachedTranslationLoader) Invalidate(entityID string) {
	_ = c.driver.Delete(context.Background(), entityID)
}
