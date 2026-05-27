package cache

import (
	"context"
	"sync"
	"time"

	"github.com/maypok86/otter"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/ports"
	"golang.org/x/sync/singleflight"
)

type otterDriver struct {
	cache  otter.CacheWithVariableTTL[string, map[string]domain.Translations]
	group  singleflight.Group
	mu     sync.Mutex
	delGen map[string]uint64
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
		cache:  cache,
		delGen: make(map[string]uint64),
	}, nil
}

func (d *otterDriver) Load(
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func(context.Context) (map[string]domain.Translations, error),
) (map[string]domain.Translations, error) {
	// Fast path: cache hit (otter is thread-safe, no lock needed).
	if val, ok := d.cache.Get(key); ok {
		return val, nil
	}

	// Capture generation before entering the singleflight group so the store
	// guard can detect a Delete that races between loader() returning and Set
	// executing.
	d.mu.Lock()
	genAtStart := d.delGen[key]
	d.mu.Unlock()

	val, err, _ := d.group.Do(key, func() (any, error) {
		// Re-check: a prior winner of this flight may have populated the cache.
		if v, ok := d.cache.Get(key); ok {
			return v, nil
		}

		result, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		if result != nil {
			// Only store if no Delete has fired since we captured genAtStart.
			d.mu.Lock()
			currentGen := d.delGen[key]
			d.mu.Unlock()
			if currentGen == genAtStart {
				d.cache.Set(key, result, ttl)
			}
		}

		return result, nil
	})
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	return val.(map[string]domain.Translations), nil
}

func (d *otterDriver) Delete(ctx context.Context, key string) error {
	// Bump generation first (under mutex) so any in-flight Do call that reads
	// currentGen after loader() returns will see the incremented value and
	// skip the Set.
	d.mu.Lock()
	d.delGen[key]++
	d.mu.Unlock()
	// Forget causes the next Do for this key to start a fresh call rather than
	// joining the stale in-flight one.
	d.group.Forget(key)
	d.cache.Delete(key)
	return nil
}
