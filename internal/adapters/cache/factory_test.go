package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDriver(t *testing.T) {
	t.Run("Default to otter driver (empty string)", func(t *testing.T) {
		cfg := Config{
			Driver:   "",
			TTL:      5 * time.Minute,
			Capacity: 100,
		}
		driver, err := NewDriver(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, driver)
	})

	t.Run("memory driver returns error", func(t *testing.T) {
		cfg := Config{Driver: "memory"}
		_, err := NewDriver(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown cache driver")
	})

	t.Run("Select otter driver", func(t *testing.T) {
		cfg := Config{
			Driver:   "otter",
			Capacity: 100,
			TTL:      1 * time.Hour,
		}
		driver, err := NewDriver(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, driver)
	})

	t.Run("Invalid driver selection", func(t *testing.T) {
		cfg := Config{
			Driver: "invalid",
		}
		_, err := NewDriver(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown cache driver")
	})
}
