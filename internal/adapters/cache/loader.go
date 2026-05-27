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

func (c *CachedTranslationLoader) BulkLoad(ctx context.Context, entityIDs []string, locales []string) (map[string]domain.Translations, error) {
	results := make(map[string]domain.Translations)

	for _, id := range entityIDs {
		entityID := id // capture for closure

		cachedMap, err := c.driver.Load(ctx, entityID, c.ttl, func(lctx context.Context) (map[string]domain.Translations, error) {
			fresh, err := c.underlying.BulkLoad(lctx, []string{entityID}, locales)
			if err != nil {
				return nil, err
			}
			trans := fresh[entityID]
			grouped := make(map[string]domain.Translations)
			for _, t := range trans {
				grouped[t.Locale] = append(grouped[t.Locale], t)
			}
			return grouped, nil
		})
		if err != nil {
			return nil, err
		}
		if cachedMap == nil {
			continue
		}

		// Verify all requested locales are present; if any are missing, evict and reload.
		allPresent := true
		for _, locale := range locales {
			if _, ok := cachedMap[locale]; !ok {
				allPresent = false
				break
			}
		}
		if !allPresent {
			if err := c.driver.Delete(ctx, entityID); err != nil {
				return nil, err
			}
			cachedMap, err = c.driver.Load(ctx, entityID, c.ttl, func(lctx context.Context) (map[string]domain.Translations, error) {
				fresh, err := c.underlying.BulkLoad(lctx, []string{entityID}, locales)
				if err != nil {
					return nil, err
				}
				trans := fresh[entityID]
				grouped := make(map[string]domain.Translations)
				for _, t := range trans {
					grouped[t.Locale] = append(grouped[t.Locale], t)
				}
				return grouped, nil
			})
			if err != nil {
				return nil, err
			}
		}

		// Flatten locale buckets into a single Translations slice.
		var entityTrans domain.Translations
		for _, locale := range locales {
			entityTrans = append(entityTrans, cachedMap[locale]...)
		}
		results[entityID] = entityTrans
	}

	return results, nil
}

func (c *CachedTranslationLoader) Invalidate(entityID string) error {
	return c.driver.Delete(context.Background(), entityID)
}
