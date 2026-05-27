//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rikeshs/translationloader/internal/adapters/cache"
	"github.com/rikeshs/translationloader/internal/adapters/repository"
	"github.com/rikeshs/translationloader/internal/app"
	"github.com/rikeshs/translationloader/internal/core/services"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFeatures(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %s", err)
	}
	defer pool.Close()

	pgLoader := repository.NewPostgresTranslationLoader(pool)
	driver, _ := cache.NewDriver(cache.Config{Driver: os.Getenv("CACHE_DRIVER")})
	cachedLoader := cache.NewCachedTranslationLoader(pgLoader, driver, 1*time.Minute)
	productRepo := repository.NewPostgresProductRepository(pool)
	docBuilder := services.NewDocumentBuilder(cachedLoader)
	handler := app.NewSyncHandler(productRepo, docBuilder, []string{"en", "th"})

	tctx := &testContext{
		db:          pool,
		pgContainer: pgContainer,
		loader:      cachedLoader,
		handler:     handler,
	}

	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			InitializeScenario(sc, tctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"sync.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
