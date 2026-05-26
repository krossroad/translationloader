package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rikeshs/translationloader/internal/adapters/cache"
	"github.com/rikeshs/translationloader/internal/adapters/repository"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/domain/dto"
	"github.com/rikeshs/translationloader/internal/core/ports"
	"github.com/rikeshs/translationloader/internal/core/services"
)

type AppConfig struct {
	DBDSN   string
	Cache   cache.Config
	Locales []string
}

type SyncApplication struct {
	pool        *pgxpool.Pool
	syncHandler *SyncHandler
}

func NewSyncApplication(ctx context.Context, cfg AppConfig) (*SyncApplication, error) {
	app := &SyncApplication{}

	if err := app.initDB(ctx, cfg.DBDSN); err != nil {
		return nil, err
	}

	cacheDriver, err := app.initDrivers(cfg.Cache)
	if err != nil {
		app.Close()
		return nil, err
	}

	loader, productRepo := app.initRepositories(cacheDriver, cfg.Cache.TTL)
	app.initServices(loader, productRepo, cfg.Locales)

	return app, nil
}

func (a *SyncApplication) initDB(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("could not ping database: %w", err)
	}

	a.pool = pool
	return nil
}

func (a *SyncApplication) initDrivers(cfg cache.Config) (ports.CacheDriver, error) {
	return cache.NewDriver(cfg)
}

func (a *SyncApplication) initRepositories(cacheDriver ports.CacheDriver, cacheTTL time.Duration) (ports.TranslationLoader, ports.ProductRepository) {
	pgTranslationLoader := repository.NewPostgresTranslationLoader(a.pool)
	cachedTranslationLoader := cache.NewCachedTranslationLoader(pgTranslationLoader, cacheDriver, cacheTTL)
	productRepo := repository.NewPostgresProductRepository(a.pool)
	return cachedTranslationLoader, productRepo
}

func (a *SyncApplication) initServices(loader ports.TranslationLoader, productRepo ports.ProductRepository, locales []string) {
	docBuilder := services.NewDocumentBuilder(loader)
	a.syncHandler = NewSyncHandler(productRepo, docBuilder, locales)
}

func (a *SyncApplication) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}

func (a *SyncApplication) RunSync(ctx context.Context, productIDs []string) ([]dto.ElasticsearchDocument, error) {
	var results []dto.ElasticsearchDocument

	for _, id := range productIDs {
		domainDoc, err := a.syncHandler.SyncProduct(ctx, id)
		if err != nil {
			log.Printf("Error syncing product %s: %v", id, err)
			continue
		}

		results = append(results, mapToDTO(domainDoc))
	}

	return results, nil
}

func mapToDTO(d domain.ProductDocument) dto.ElasticsearchDocument {
	productNames := make([]dto.ProductName, len(d.ProductName))
	for i, pn := range d.ProductName {
		productNames[i] = dto.ProductName{
			Locale: pn.Locale,
			Data:   pn.Data,
		}
	}

	return dto.ElasticsearchDocument{
		UUID:       d.UUID,
		SKU:        d.SKU,
		PartNumber: d.PartNumber,
		Brand: dto.Brand{
			Code: d.Brand.Code,
			Label: dto.Label{
				En: d.Brand.Label.En,
				Th: d.Brand.Label.Th,
			},
		},
		ProductName: productNames,
		OilGrade: dto.Property{
			Code: d.OilGrade.Code,
			Label: dto.Label{
				En: d.OilGrade.Label.En,
				Th: d.OilGrade.Label.Th,
			},
		},
		Attributes: d.Attributes,
	}
}
