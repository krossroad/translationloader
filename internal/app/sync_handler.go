package app

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/ports"
)

type SyncHandler struct {
	productRepo ports.ProductRepository
	docBuilder  ports.DocumentBuilder
	locales     []string
}

func NewSyncHandler(productRepo ports.ProductRepository, docBuilder ports.DocumentBuilder, locales []string) *SyncHandler {
	l := locales
	// sort.Slice(l)
	return &SyncHandler{
		productRepo: productRepo,
		docBuilder:  docBuilder,
		locales:     l,
	}
}

func (h *SyncHandler) SyncProduct(ctx context.Context, id string) (domain.ProductDocument, error) {
	var (
		p     domain.Product
		attrs []domain.Attribute
		specs []domain.ProductSpecification
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		p, err = h.productRepo.GetProduct(gctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch product %s: %w", id, err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		attrs, err = h.productRepo.GetAttributesByProductID(gctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch attributes for product %s: %w", id, err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		specs, err = h.productRepo.GetSpecificationsByProductID(gctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch specifications for product %s: %w", id, err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return domain.ProductDocument{}, err
	}

	return h.docBuilder.BuildProductDocument(ctx, p, attrs, specs, h.locales)
}
