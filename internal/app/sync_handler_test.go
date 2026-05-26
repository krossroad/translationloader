package app

import (
	"context"
	"testing"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSyncHandler_SyncProduct(t *testing.T) {
	ctx := context.Background()
	id := "p-1"
	locales := []string{"en", "th"}

	product := domain.Product{ID: id, SKU: "SKU-1"}
	attrs := []domain.Attribute{{ID: "a-1", Code: "oil_grade"}}
	specs := []domain.ProductSpecification{{ID: "s-1", ProductID: id, AttributeID: "a-1", Value: "5w30"}}
	expectedDoc := domain.ProductDocument{UUID: id, SKU: "SKU-1"}

	tests := []struct {
		name          string
		setupMocks    func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder)
		expectedDoc   domain.ProductDocument
		expectedError string
	}{
		{
			name: "success",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", ctx, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", ctx, id).Return(attrs, nil)
				mRepo.On("GetSpecificationsByProductID", ctx, id).Return(specs, nil)
				mBuilder.On("BuildProductDocument", ctx, product, attrs, specs, locales).Return(expectedDoc, nil)
			},
			expectedDoc: expectedDoc,
		},
		{
			name: "product not found",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", ctx, id).Return(domain.Product{}, assert.AnError)
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name: "attributes load error",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", ctx, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", ctx, id).Return([]domain.Attribute{}, assert.AnError)
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name: "specifications load error",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", ctx, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", ctx, id).Return(attrs, nil)
				mRepo.On("GetSpecificationsByProductID", ctx, id).Return([]domain.ProductSpecification{}, assert.AnError)
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name: "builder error",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", ctx, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", ctx, id).Return(attrs, nil)
				mRepo.On("GetSpecificationsByProductID", ctx, id).Return(specs, nil)
				mBuilder.On("BuildProductDocument", ctx, product, attrs, specs, locales).Return(domain.ProductDocument{}, assert.AnError)
			},
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewProductRepository(t)
			mockBuilder := mocks.NewDocumentBuilder(t)
			handler := NewSyncHandler(mockRepo, mockBuilder, locales)

			tt.setupMocks(mockRepo, mockBuilder)

			doc, err := handler.SyncProduct(ctx, id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDoc, doc)
			}
		})
	}
}
