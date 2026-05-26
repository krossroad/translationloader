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
		cachedMap, err := c.driver.Get(ctx, id)
		if err != nil || cachedMap == nil {
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

		// Flatten nested map back to []Translation for the BulkLoad result
		var entityTrans []domain.Translation
		for _, locale := range locales {
			for _, t := range cachedMap[locale] {
				entityTrans = append(entityTrans, t)
			}
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

			// Store in cache grouped by locale then field name
			grouped := make(map[string]domain.Translations)
			for _, t := range trans {
				if grouped[t.Locale] == nil {
					grouped[t.Locale] = make(domain.Translations)
				}
				grouped[t.Locale][t.FieldName] = t
			}
			_ = c.driver.Set(ctx, id, grouped, c.ttl)
		}
	}

	return results, nil
}

func (c *CachedTranslationLoader) Invalidate(entityID string) error {
	return c.driver.Delete(context.Background(), entityID)
}
