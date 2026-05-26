package ports

import (
	"context"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name TranslationLoader --output ../../../test/mocks --outpkg mocks --case underscore
type TranslationLoader interface {
	// BulkLoad fetches translations for a batch of entity IDs and specific locales.
	// Returns a map where key is EntityID and value is a slice of translations for that entity.
	BulkLoad(ctx context.Context, entityIDs []string, locales []string) (map[string][]domain.Translation, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name CacheDriver --output ../../../test/mocks --outpkg mocks --case underscore
type CacheDriver interface {
	// Get retrieves all cached translations for an entity.
	// The return value is a map where the key is the locale.
	// Returns the map, a boolean indicating if the key was found, and any error.
	Get(ctx context.Context, key string) (map[string][]domain.Translation, bool, error)

	// Set stores translations for an entity with a specific TTL.
	// The value is a map where the key is the locale.
	Set(ctx context.Context, key string, value map[string][]domain.Translation, ttl time.Duration) error

	// Delete removes all translations for a specific entity.
	Delete(ctx context.Context, key string) error
}
