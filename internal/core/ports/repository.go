package ports

import (
	"context"
	"github.com/rikeshs/translationloader/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name ProductRepository --output ../../../test/mocks --outpkg mocks --case underscore
type ProductRepository interface {
	GetProduct(ctx context.Context, id string) (domain.Product, error)
	GetAttributesByProductID(ctx context.Context, productID string) ([]domain.Attribute, error)
	GetSpecificationsByProductID(ctx context.Context, productID string) ([]domain.ProductSpecification, error)
}
