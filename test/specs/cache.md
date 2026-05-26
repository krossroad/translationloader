# Specification: Pluggable Cache Architecture

## 1. Overview
The `TranslationLoader` needs a robust caching layer to reduce the load on the underlying repository (e.g., Postgres). The new architecture uses an entity-centric approach where all translations for a given entity are stored in a single cache entry, enabling efficient $O(1)$ invalidation and improved hit rates for locale subsets.

## 2. Interface Specification: `CacheDriver`
The `CacheDriver` is a low-level interface that handles basic key-value operations.

```go
package ports

import (
	"context"
	"time"
	"github.com/rikeshs/translationloader/internal/core/domain"
)

type CacheDriver interface {
	// Get retrieves all cached translations for an entity.
	// The return value is a map where the key is the locale.
	// Returns the map, a boolean indicating if it was a hit, and any error.
	Get(ctx context.Context, key string) (map[string][]domain.Translation, bool, error)
	
	// Set stores translations for an entity with a specific TTL.
	// The value is a map where the key is the locale.
	Set(ctx context.Context, key string, value map[string][]domain.Translation, ttl time.Duration) error
	
	// Delete removes all translations for a specific entity.
	Delete(ctx context.Context, key string) error
}
```

## 3. Orchestrator Specification: `CachedTranslationLoader`
The `CachedTranslationLoader` acts as an orchestrator that implements the `ports.TranslationLoader` interface. It manages the caching logic using the `entityID` as the primary key.

### 3.1. Entity-Centric Storage
- **Cache Key**: The `entityID`.
- **Cache Value**: `map[string][]domain.Translation` where the key is the `locale`.
- **O(1) Invalidation**: Since all locales for an entity are stored under a single key, invalidation is a direct $O(1)$ call to `driver.Delete(entityID)`. No secondary index is required.

### 3.2. Logical Workflow
- **BulkLoad(entityIDs, locales)**:
  For each requested `entityID`:
  1. `Get(entityID)` from the `CacheDriver`.
  2. If hit:
     - Check if **all** requested `locales` are present as keys in the returned map.
     - If all locales are present: Add translations for the requested locales to the final results.
     - If any requested locale is missing (partial hit): Proceed to step 3.
  3. If miss OR partial hit:
     - Load translations from the `underlying` repository for that entity, requesting all required `locales`.
     - Store the new map in the `CacheDriver` using `Set(entityID, ...)`. This overwrites any existing entry.
     - Add results to final output.

- **Invalidate(entityID)**:
  1. Call `CacheDriver.Delete(entityID)`. This is a direct $O(1)$ operation.

## 4. Driver Implementation: Otter
Otter is a high-performance, lockless Go cache using the S3-FIFO algorithm.

### 4.1. Configuration
- **Capacity**: Maximum number of items the cache can hold.
- **Algorithm**: S3-FIFO (handled by Otter internally).
- **TTL**: Default expiration time for items.

### 4.2. Implementation Details
The Otter driver will wrap `otter.Cache[string, map[string][]domain.Translation]`.

## 5. System Configuration
The application will select the driver based on environment variables.

| Variable | Description | Default |
|----------|-------------|---------|
| `CACHE_DRIVER` | Driver type: `memory` (simple map) or `otter` | `memory` |
| `CACHE_TTL` | Default TTL (e.g., `1h`, `30m`) | `1h` |
| `CACHE_OTTER_CAPACITY` | Max items for Otter | `10000` |

## 6. Acceptance Criteria

### Scenario: Successful cache hit for locale subset
**Given** entity "PROD-1" has translations for ["en", "fr", "de"] cached in a single entry
**When** a request is made for "PROD-1" translations in ["en", "de"]
**Then** the results for "en" and "de" should be returned from the cache
**And** the underlying repository should NOT be called.

### Scenario: Partial miss triggers reload and overwrite
**Given** entity "PROD-1" has translations for ["en", "fr"] cached
**When** a request is made for "PROD-1" translations in ["en", "es"]
**Then** the cache should identify that "es" is missing (partial miss)
**And** it should load "en" and "es" from the underlying repository
**And** it should store the new map {"en": [...], "es": [...]} in the cache, overwriting the old entry.

### Scenario: O(1) Invalidation of an entity
**Given** entity "PROD-3" has translations cached for multiple locales
**When** `Invalidate("PROD-3")` is called
**Then** the entry for "PROD-3" should be immediately deleted from the driver using a single $O(1)$ operation
**And** no linear scan or secondary index lookups should occur.

### Scenario: Memory safety with bounded cache
**Given** `CACHE_OTTER_CAPACITY` is set to 100
**When** 101 unique entities are loaded into the cache
**Then** the cache should evict the oldest/least-used item according to S3-FIFO
**And** memory usage should remain stable.
