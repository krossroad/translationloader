// Package app provides the application orchestration logic.
package app

import (
	"context"
	"fmt"
	"log"
	"maps"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rikeshs/translationloader/internal/adapters/cache"
	"github.com/rikeshs/translationloader/internal/adapters/repository"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/domain/dto"
	"github.com/rikeshs/translationloader/internal/core/services"
)

// Config holds the configuration for the SyncApplication.
type Config struct {
	DBDSN   string
	Cache   cache.Config
	Locales []string
}

// SyncApplication represents the main application service.
type SyncApplication struct {
	syncHandler *SyncHandler
	pgPool      *pgxpool.Pool
}

// NewSyncApplication initializes a new SyncApplication instance with required dependencies.
func NewSyncApplication(ctx context.Context, cfg Config) (*SyncApplication, error) {
	pool, err := initDB(ctx, cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	cacheDriver, err := cache.NewDriver(cfg.Cache)
	if err != nil {
		pool.Close()
		return nil, err
	}

	pgTranslationLoader := repository.NewPostgresTranslationLoader(pool)
	cachedTranslationLoader := cache.NewCachedTranslationLoader(pgTranslationLoader, cacheDriver, cfg.Cache.TTL)
	productRepo := repository.NewPostgresProductRepository(pool)

	docBuilder := services.NewDocumentBuilder(cachedTranslationLoader)
	handler := NewSyncHandler(productRepo, docBuilder, cfg.Locales)

	return &SyncApplication{syncHandler: handler, pgPool: pool}, nil
}

// NewSyncApplicationFromHandler creates a SyncApplication from an already-wired handler.
// Use this when adapter construction is handled outside the app layer (e.g., in cmd/).
func NewSyncApplicationFromHandler(handler *SyncHandler) *SyncApplication {
	return &SyncApplication{syncHandler: handler}
}

func initDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	return pool, nil
}

// Close closes the application resources, such as database pools.
func (a *SyncApplication) Close() {
	if a.pgPool != nil {
		a.pgPool.Close()
	}
}

// BuildProductDocument builds a list of Elasticsearch documents for the given product IDs.
func (a *SyncApplication) BuildProductDocument(ctx context.Context, productIDs []string) ([]dto.ElasticsearchDocument, error) {
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

	brandLabel := dto.Label(maps.Clone(d.Brand.Label))
	oilGradeLabel := dto.Label(maps.Clone(d.OilGrade.Label))

	return dto.ElasticsearchDocument{
		UUID:       d.UUID,
		SKU:        d.SKU,
		PartNumber: d.PartNumber,
		Brand: dto.Brand{
			Code:  d.Brand.Code,
			Label: brandLabel,
		},
		ProductName: productNames,
		OilGrade: dto.Property{
			Code:  d.OilGrade.Code,
			Label: oilGradeLabel,
		},
		Attributes: d.Attributes,
	}
}
