package ports

import (
	"context"
	"github.com/rikeshs/translationloader/internal/core/domain"
)

type ProductRepository interface {
	GetProduct(ctx context.Context, id string) (domain.Product, error)
	GetAttributesByProductID(ctx context.Context, productID string) ([]domain.Attribute, error)
	GetSpecificationsByProductID(ctx context.Context, productID string) ([]domain.ProductSpecification, error)
}
