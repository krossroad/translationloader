# Translation Loader

A robust, performant Go service designed to load and sync translations, utilizing a lite hexagonal architecture for maintainability and the Otter caching library for high-speed access.

## How to Run

### Prerequisites
- Go 1.23+
- Docker & Docker Compose

### Setup & Infrastructure
1. Start the required infrastructure (PostgreSQL):
   ```bash
   make docker-up
   ```

2. Load data fixtures:
   ```bash
   make load-fixtures
   ```

3. Run the application:
   ```bash
   # Ensure DATABASE_URL is set if using non-default configuration
   make run
   ```

### Development
- **Run tests:** `make test`
- **Run linter:** `make lint`
- **Generate mocks:** `make generate-mocks`

## Design Decisions

### Lite Hexagonal Architecture
The project follows a simplified hexagonal (ports and adapters) architecture to decouple business logic from external systems:
- `internal/core`: Contains the domain models and business services (`ports` and `services`).
- `internal/app`: Implements the application layer, orchestration, and handlers.
- `internal/adapters`: Contains external implementation details, such as PostgreSQL repositories and caching drivers.

### Entity-Centric Cache
To ensure low latency, translations are cached using an entity-centric approach. We utilize the [Otter](https://github.com/maypok86/otter) caching library, which provides high-concurrency performance and S3-FIFO eviction policies, ideal for read-heavy translation workloads.

## Future Improvements

- **Configuration Management:** Migrate from environment variables and simple config to a robust configuration management solution (e.g., [Viper](https://github.com/spf13/viper) or [Cleanenv](https://github.com/caarlos0/env)) for better hierarchical config loading.
- **Observability:** Integrate [OpenTelemetry](https://opentelemetry.io/) for distributed tracing and implement Prometheus metrics to monitor sync performance and cache hit/miss ratios.
- **Robust Data Seeding/Migrations:** Implement a more automated migration process and robust data seeding mechanism directly integrated into the CI/CD pipeline or the application startup sequence for ephemeral environments.
