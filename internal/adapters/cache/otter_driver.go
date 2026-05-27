package cache

import (
	"context"
	"time"

	"github.com/maypok86/otter"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/ports"
)

type otterDriver struct {
	cache otter.CacheWithVariableTTL[string, map[string]domain.Translations]
}

func NewOtterDriver(capacity int, defaultTTL time.Duration) (ports.CacheDriver, error) {
	cache, err := otter.MustBuilder[string, map[string]domain.Translations](capacity).
		CollectStats().
		Cost(func(key string, value map[string]domain.Translations) uint32 {
			return 1
		}).
		WithVariableTTL().
		Build()
	if err != nil {
		return nil, err
	}

	return &otterDriver{
		cache: cache,
	}, nil
}

func (d *otterDriver) Load(
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func(context.Context) (map[string]domain.Translations, error),
) (map[string]domain.Translations, error) {
	// Fast path: cache hit.
	if val, ok := d.cache.Get(key); ok {
		return val, nil
	}

	// Slow path: miss — invoke loader and store the result.
	result, err := loader(ctx)
	if err != nil {
		return nil, err
	}
	if result != nil {
		d.cache.Set(key, result, ttl)
	}
	return result, nil
}

func (d *otterDriver) Delete(ctx context.Context, key string) error {
	d.cache.Delete(key)
	return nil
}
