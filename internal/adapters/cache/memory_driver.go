package cache

import (
	"context"
	"sync"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/ports"
)

type memoryItem struct {
	value     map[string][]domain.Translation
	expiresAt time.Time
}

type memoryDriver struct {
	items map[string]memoryItem
	mu    sync.RWMutex
}

func NewMemoryDriver() ports.CacheDriver {
	return &memoryDriver{
		items: make(map[string]memoryItem),
	}
}

func (d *memoryDriver) Get(ctx context.Context, key string) (map[string][]domain.Translation, bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	item, ok := d.items[key]
	if !ok {
		return nil, false, nil
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return nil, false, nil
	}

	return item.value, true, nil
}

func (d *memoryDriver) Set(ctx context.Context, key string, value map[string][]domain.Translation, ttl time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	d.items[key] = memoryItem{
		value:     value,
		expiresAt: expiresAt,
	}
	return nil
}

func (d *memoryDriver) Delete(ctx context.Context, key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.items, key)
	return nil
}
