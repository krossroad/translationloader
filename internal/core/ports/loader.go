package ports

import (
	"context"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
)

type TranslationLoader interface {
	// BulkLoad fetches translations for a batch of entity IDs and specific locales.
	// Returns a map where key is EntityID and value is a slice of translations for that entity.
	BulkLoad(ctx context.Context, entityIDs []string, locales []string) (map[string]domain.Translations, error)
}

type CacheDriver interface {
	// Load returns the cached value for key, or calls loader on a miss and
	// stores the result. The driver serialises miss→fetch→store under a
	// per-key lock so a concurrent Delete cannot race with the store.
	Load(ctx context.Context, key string, ttl time.Duration,
		loader func(context.Context) (map[string]domain.Translations, error),
	) (map[string]domain.Translations, error)

	// Delete removes the cached value for key (entity-level, O(1)).
	Delete(ctx context.Context, key string) error
}
