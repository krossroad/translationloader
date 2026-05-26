package app

import (
	"context"
	"testing"

	"github.com/rikeshs/translationloader/internal/adapters/cache"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMapToDTO(t *testing.T) {
	doc := domain.ProductDocument{
		UUID: "p-1",
		SKU:  "SKU-1",
		ProductName: []domain.ProductName{
			{Locale: "en", Data: "Oil 1L"},
		},
		Brand: domain.BrandInfo{
			Code: "castrol",
			Label: domain.Label{
				En: "Castrol",
				Th: "คาสตรอล",
			},
		},
		OilGrade: domain.Property{
			Code: "5w30",
			Label: domain.Label{
				En: "5W-30",
				Th: "5W-30",
			},
		},
		Attributes: map[string]string{
			"viscosity": "5w30",
		},
	}

	dto := mapToDTO(doc)

	assert.Equal(t, doc.UUID, dto.UUID)
	assert.Equal(t, doc.SKU, dto.SKU)
	assert.Equal(t, len(doc.ProductName), len(dto.ProductName))
	assert.Equal(t, doc.ProductName[0].Locale, dto.ProductName[0].Locale)
	assert.Equal(t, doc.ProductName[0].Data, dto.ProductName[0].Data)
	assert.Equal(t, doc.Brand.Code, dto.Brand.Code)
	assert.Equal(t, doc.Brand.Label.En, dto.Brand.Label.En)
	assert.Equal(t, doc.OilGrade.Code, dto.OilGrade.Code)
	assert.Equal(t, doc.Attributes["viscosity"], dto.Attributes["viscosity"])
}

func TestSyncApplication_RunSync(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewProductRepository(t)
	mockBuilder := mocks.NewDocumentBuilder(t)
	handlerLocales := []string{"en", "th"}
	handler := NewSyncHandler(mockRepo, mockBuilder, handlerLocales)

	app := &SyncApplication{
		syncHandler: handler,
	}

	productIDs := []string{"p-1", "p-2"}

	t.Run("success all", func(t *testing.T) {
		doc1 := domain.ProductDocument{UUID: "p-1", SKU: "S1"}
		doc2 := domain.ProductDocument{UUID: "p-2", SKU: "S2"}

		mockRepo.On("GetProduct", ctx, "p-1").Return(domain.Product{ID: "p-1", SKU: "S1"}, nil).Once()
		mockRepo.On("GetAttributesByProductID", ctx, "p-1").Return([]domain.Attribute{}, nil).Once()
		mockRepo.On("GetSpecificationsByProductID", ctx, "p-1").Return([]domain.ProductSpecification{}, nil).Once()
		mockBuilder.On("BuildProductDocument", ctx, mock.Anything, mock.Anything, mock.Anything, handlerLocales).Return(doc1, nil).Once()

		mockRepo.On("GetProduct", ctx, "p-2").Return(domain.Product{ID: "p-2", SKU: "S2"}, nil).Once()
		mockRepo.On("GetAttributesByProductID", ctx, "p-2").Return([]domain.Attribute{}, nil).Once()
		mockRepo.On("GetSpecificationsByProductID", ctx, "p-2").Return([]domain.ProductSpecification{}, nil).Once()
		mockBuilder.On("BuildProductDocument", ctx, mock.Anything, mock.Anything, mock.Anything, handlerLocales).Return(doc2, nil).Once()

		res, err := app.RunSync(ctx, productIDs)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "p-1", res[0].UUID)
		assert.Equal(t, "p-2", res[1].UUID)
	})

	t.Run("partial failure", func(t *testing.T) {
		mockRepo.On("GetProduct", ctx, "p-1").Return(domain.Product{}, assert.AnError).Once()
		
		doc2 := domain.ProductDocument{UUID: "p-2", SKU: "S2"}
		mockRepo.On("GetProduct", ctx, "p-2").Return(domain.Product{ID: "p-2", SKU: "S2"}, nil).Once()
		mockRepo.On("GetAttributesByProductID", ctx, "p-2").Return([]domain.Attribute{}, nil).Once()
		mockRepo.On("GetSpecificationsByProductID", ctx, "p-2").Return([]domain.ProductSpecification{}, nil).Once()
		mockBuilder.On("BuildProductDocument", ctx, mock.Anything, mock.Anything, mock.Anything, handlerLocales).Return(doc2, nil).Once()

		res, err := app.RunSync(ctx, productIDs)
		assert.NoError(t, err) // Should continue on per-product error
		assert.Len(t, res, 1)
		assert.Equal(t, "p-2", res[0].UUID)
	})
}

func TestNewSyncApplication(t *testing.T) {
	ctx := context.Background()
	cfg := AppConfig{
		DBDSN: "invalid-dsn",
		Cache: cache.Config{
			Driver: "memory",
		},
	}

	app, err := NewSyncApplication(ctx, cfg)
	assert.Error(t, err)
	assert.Nil(t, app)
}

func TestMapToDTO_Empty(t *testing.T) {
	doc := domain.ProductDocument{}
	dto := mapToDTO(doc)

	assert.Equal(t, "", dto.UUID)
	assert.Equal(t, "", dto.SKU)
	assert.Empty(t, dto.ProductName)
	assert.Empty(t, dto.Attributes)
}

func TestSyncApplication_Close(t *testing.T) {
	// Simple test to ensure Close doesn't panic even if pool is nil
	app := &SyncApplication{}
	assert.NotPanics(t, func() {
		app.Close()
	})
}
