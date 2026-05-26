package ports

import (
	"context"

	"github.com/rikeshs/translationloader/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.6 --name DocumentBuilder --output ../../../test/mocks --outpkg mocks --case underscore
type DocumentBuilder interface {
	BuildProductDocument(ctx context.Context, p domain.Product, attrs []domain.Attribute, specs []domain.ProductSpecification, locales []string) (domain.ProductDocument, error)
}
