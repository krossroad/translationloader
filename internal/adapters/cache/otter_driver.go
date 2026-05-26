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

func (d *otterDriver) Get(ctx context.Context, key string) (map[string]domain.Translations, error) {
	val, ok := d.cache.Get(key)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (d *otterDriver) Set(ctx context.Context, key string, value map[string]domain.Translations, ttl time.Duration) error {
	d.cache.Set(key, value, ttl)
	return nil
}

func (d *otterDriver) Delete(ctx context.Context, key string) error {
	d.cache.Delete(key)
	return nil
}
