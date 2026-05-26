package ports

import (
	"context"

	"github.com/rikeshs/translationloader/internal/core/domain"
)

type DocumentBuilder interface {
	BuildProductDocument(ctx context.Context, p domain.Product, attrs []domain.Attribute, specs []domain.ProductSpecification, locales []string) (domain.ProductDocument, error)
}
