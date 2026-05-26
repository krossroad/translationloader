package cache

import (
	"fmt"
	"time"

	"github.com/rikeshs/translationloader/internal/core/ports"
)

type Config struct {
	Driver   string
	TTL      time.Duration
	Capacity int
}

func NewDriver(cfg Config) (ports.CacheDriver, error) {
	if cfg.TTL == 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.Capacity == 0 {
		cfg.Capacity = 1000
	}

	switch cfg.Driver {
	case "otter":
		return NewOtterDriver(cfg.Capacity, cfg.TTL)
	case "memory", "":
		return NewMemoryDriver(), nil
	default:
		return nil, fmt.Errorf("unknown cache driver: %s", cfg.Driver)
	}
}
