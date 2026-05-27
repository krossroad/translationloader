//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rikeshs/translationloader/internal/adapters/cache"
	"github.com/rikeshs/translationloader/internal/adapters/repository"
	"github.com/rikeshs/translationloader/internal/app"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/services"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type testContext struct {
	db           *pgxpool.Pool
	pgContainer  *postgres.PostgresContainer
	loader       *cache.CachedTranslationLoader
	handler      *app.SyncHandler
	lastDocument domain.ProductDocument
	lastError    error
}

func (c *testContext) aCleanPostgreSQLDatabaseWithTheProductSchema(ctx context.Context) error {
	// Re-initialize loader and handler to clear cache
	pgLoader := repository.NewPostgresTranslationLoader(c.db)
	driver, _ := cache.NewDriver(cache.Config{Driver: os.Getenv("CACHE_DRIVER")})
	c.loader = cache.NewCachedTranslationLoader(pgLoader, driver, 1*time.Minute)
	productRepo := repository.NewPostgresProductRepository(c.db)
	docBuilder := services.NewDocumentBuilder(c.loader)
	c.handler = app.NewSyncHandler(productRepo, docBuilder, []string{"en", "th"})

	queries := []string{
		"DROP TABLE IF EXISTS translation",
		"DROP TABLE IF EXISTS product_specification",
		"DROP TABLE IF EXISTS attribute",
		"DROP TABLE IF EXISTS product",
		"CREATE TABLE product (id UUID PRIMARY KEY, sku VARCHAR, part_number VARCHAR, brand VARCHAR, category_id UUID)",
		"CREATE TABLE attribute (id UUID PRIMARY KEY, code VARCHAR, metric_unit VARCHAR)",
		"CREATE TABLE product_specification (id UUID PRIMARY KEY, product_id UUID REFERENCES product(id), attribute_id UUID REFERENCES attribute(id), value VARCHAR)",
		"CREATE TABLE translation (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), entity_type VARCHAR, entity_id VARCHAR, locale VARCHAR, field_name VARCHAR, field_value TEXT, updated_at TIMESTAMPTZ DEFAULT NOW())",
	}

	for _, q := range queries {
		if _, err := c.db.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (c *testContext) theFollowingProductsExist(ctx context.Context, table *godog.Table) error {
	for i := 1; i < len(table.Rows); i++ {
		row := table.Rows[i]
		_, err := c.db.Exec(ctx, "INSERT INTO product (id, sku, part_number, brand) VALUES ($1, $2, $3, $4)",
			row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *testContext) theFollowingAttributesExist(ctx context.Context, table *godog.Table) error {
	for i := 1; i < len(table.Rows); i++ {
		row := table.Rows[i]
		_, err := c.db.Exec(ctx, "INSERT INTO attribute (id, code, metric_unit) VALUES ($1, $2, $3)",
			row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *testContext) theFollowingSpecificationsExist(ctx context.Context, table *godog.Table) error {
	for i := 1; i < len(table.Rows); i++ {
		row := table.Rows[i]
		_, err := c.db.Exec(ctx, "INSERT INTO product_specification (id, product_id, attribute_id, value) VALUES ($1, $2, $3, $4)",
			row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *testContext) theFollowingTranslationsExist(ctx context.Context, table *godog.Table) error {
	for i := 1; i < len(table.Rows); i++ {
		row := table.Rows[i]
		_, err := c.db.Exec(ctx, "INSERT INTO translation (entity_type, entity_id, locale, field_name, field_value) VALUES ($1, $2, $3, $4, $5)",
			row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value, row.Cells[4].Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *testContext) iBuildTheDocumentForProductWithLocales(ctx context.Context, productID string, localesStr string) error {
	locales := strings.Split(localesStr, ",")

	productRepo := repository.NewPostgresProductRepository(c.db)
	docBuilder := services.NewDocumentBuilder(c.loader)
	c.handler = app.NewSyncHandler(productRepo, docBuilder, locales)

	doc, err := c.handler.SyncProduct(ctx, productID)
	c.lastError = err
	c.lastDocument = doc
	return nil
}

func (c *testContext) theDocumentBuildShouldFail() error {
	if c.lastError == nil {
		return fmt.Errorf("expected an error but got none")
	}
	return nil
}

func (c *testContext) theDocumentSKUShouldBe(sku string) error {
	if c.lastDocument.SKU != sku {
		return fmt.Errorf("expected SKU %s, got %s", sku, c.lastDocument.SKU)
	}
	return nil
}

func (c *testContext) theDocumentShouldContainTheEnglishProductName(name string) error {
	for _, n := range c.lastDocument.ProductName {
		if n.Locale == "en" && n.Data == name {
			return nil
		}
	}
	return fmt.Errorf("could not find English product name %s", name)
}

func (c *testContext) theDocumentShouldContainTheThaiProductName(name string) error {
	for _, n := range c.lastDocument.ProductName {
		if n.Locale == "th" && n.Data == name {
			return nil
		}
	}
	return fmt.Errorf("could not find Thai product name %s", name)
}

func (c *testContext) theDocumentsOil_gradeEnglishLabelShouldBe(label string) error {
	if c.lastDocument.OilGrade.Label.En != label {
		return fmt.Errorf("expected oil_grade EN label %s, got %s", label, c.lastDocument.OilGrade.Label.En)
	}
	return nil
}

func (c *testContext) theTranslationForProductLocaleFieldIsUpdatedToInTheDatabase(ctx context.Context, id, locale, field, newValue string) error {
	_, err := c.db.Exec(ctx, "UPDATE translation SET field_value = $1 WHERE entity_id = $2 AND locale = $3 AND field_name = $4",
		newValue, id, locale, field)
	return err
}

func (c *testContext) iInvalidateTheCacheForEntity(id string) error {
	c.loader.Invalidate(id)
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext, tctx *testContext) {
	ctx.Step(`^a clean PostgreSQL database with the product schema$`, tctx.aCleanPostgreSQLDatabaseWithTheProductSchema)
	ctx.Step(`^the following products exist:$`, tctx.theFollowingProductsExist)
	ctx.Step(`^the following attributes exist:$`, tctx.theFollowingAttributesExist)
	ctx.Step(`^the following specifications exist:$`, tctx.theFollowingSpecificationsExist)
	ctx.Step(`^the following translations exist:$`, tctx.theFollowingTranslationsExist)
	ctx.Step(`^I build the document for product "([^"]*)" with locales "([^"]*)"$`, tctx.iBuildTheDocumentForProductWithLocales)
	ctx.Step(`^the document build should fail$`, tctx.theDocumentBuildShouldFail)
	ctx.Step(`^the document SKU should be "([^"]*)"$`, tctx.theDocumentSKUShouldBe)
	ctx.Step(`^the document should contain the English product name "([^"]*)"$`, tctx.theDocumentShouldContainTheEnglishProductName)
	ctx.Step(`^the document should contain the Thai product name "([^"]*)"$`, tctx.theDocumentShouldContainTheThaiProductName)
	ctx.Step(`^the document\'s oil_grade English label should be "([^"]*)"$`, tctx.theDocumentsOil_gradeEnglishLabelShouldBe)
	ctx.Step(`^the translation for product "([^"]*)" \(locale "([^"]*)", field "([^"]*)"\) is updated to "([^"]*)" in the database$`, tctx.theTranslationForProductLocaleFieldIsUpdatedToInTheDatabase)
	ctx.Step(`^I invalidate the cache for entity "([^"]*)"$`, tctx.iInvalidateTheCacheForEntity)
}
