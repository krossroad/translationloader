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
	// Get retrieves cached translations for an entity keyed by locale then field name.
	// Returns nil map (not an error) on a cache miss.
	Get(ctx context.Context, key string) (map[string]domain.Translations, error)

	// Set stores translations for an entity with a specific TTL.
	// The value is a map where the outer key is locale and the inner key is field name.
	Set(ctx context.Context, key string, value map[string]domain.Translations, ttl time.Duration) error

	// Delete removes all translations for a specific entity.
	Delete(ctx context.Context, key string) error
}
