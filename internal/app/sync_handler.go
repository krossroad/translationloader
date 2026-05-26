package app

import (
	"context"
	"fmt"

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
	p, err := h.productRepo.GetProduct(ctx, id)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("failed to fetch product %s: %w", id, err)
	}

	attrs, err := h.productRepo.GetAttributesByProductID(ctx, id)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("failed to fetch attributes for product %s: %w", id, err)
	}

	specs, err := h.productRepo.GetSpecificationsByProductID(ctx, id)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("failed to fetch specifications for product %s: %w", id, err)
	}

	return h.docBuilder.BuildProductDocument(ctx, p, attrs, specs, h.locales)
}
