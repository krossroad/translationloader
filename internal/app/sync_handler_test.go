package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				mRepo.On("GetProduct", mock.Anything, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", mock.Anything, id).Return(attrs, nil)
				mRepo.On("GetSpecificationsByProductID", mock.Anything, id).Return(specs, nil)
				mBuilder.On("BuildProductDocument", ctx, product, attrs, specs, locales).Return(expectedDoc, nil)
			},
			expectedDoc: expectedDoc,
		},
		{
			name: "product not found",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", mock.Anything, id).Return(domain.Product{}, assert.AnError)
				mRepo.On("GetAttributesByProductID", mock.Anything, id).Maybe().Return([]domain.Attribute{}, nil)
				mRepo.On("GetSpecificationsByProductID", mock.Anything, id).Maybe().Return([]domain.ProductSpecification{}, nil)
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name: "attributes load error",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", mock.Anything, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", mock.Anything, id).Return([]domain.Attribute{}, assert.AnError)
				mRepo.On("GetSpecificationsByProductID", mock.Anything, id).Maybe().Return([]domain.ProductSpecification{}, nil)
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name: "specifications load error",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", mock.Anything, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", mock.Anything, id).Return(attrs, nil)
				mRepo.On("GetSpecificationsByProductID", mock.Anything, id).Return([]domain.ProductSpecification{}, assert.AnError)
			},
			expectedError: assert.AnError.Error(),
		},
		{
			name: "builder error",
			setupMocks: func(mRepo *mocks.ProductRepository, mBuilder *mocks.DocumentBuilder) {
				mRepo.On("GetProduct", mock.Anything, id).Return(product, nil)
				mRepo.On("GetAttributesByProductID", mock.Anything, id).Return(attrs, nil)
				mRepo.On("GetSpecificationsByProductID", mock.Anything, id).Return(specs, nil)
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

// TestSyncHandler_SyncProduct_Concurrent verifies that GetProduct, GetAttributesByProductID,
// and GetSpecificationsByProductID are issued concurrently rather than sequentially.
// Each mock call sleeps for 60ms. If sequential, the total elapsed time is ~180ms.
// The test asserts elapsed < 2×delay (120ms), which only holds when all three calls
// are fanned out in parallel. This test CURRENTLY FAILS because SyncProduct is sequential.
func TestSyncHandler_SyncProduct_Concurrent(t *testing.T) {
	delay := 60 * time.Millisecond
	ctx := context.Background()
	id := "p-1"

	var (
		startTimes [3]time.Time
		mu         sync.Mutex
	)

	mockRepo := mocks.NewProductRepository(t)
	mockBuilder := mocks.NewDocumentBuilder(t)

	product := domain.Product{ID: id, SKU: "SKU-1"}

	mockRepo.On("GetProduct", mock.Anything, id).
		Run(func(args mock.Arguments) {
			mu.Lock()
			startTimes[0] = time.Now()
			mu.Unlock()
			time.Sleep(delay)
		}).
		Return(product, nil).Once()

	mockRepo.On("GetAttributesByProductID", mock.Anything, id).
		Run(func(args mock.Arguments) {
			mu.Lock()
			startTimes[1] = time.Now()
			mu.Unlock()
			time.Sleep(delay)
		}).
		Return([]domain.Attribute{}, nil).Once()

	mockRepo.On("GetSpecificationsByProductID", mock.Anything, id).
		Run(func(args mock.Arguments) {
			mu.Lock()
			startTimes[2] = time.Now()
			mu.Unlock()
			time.Sleep(delay)
		}).
		Return([]domain.ProductSpecification{}, nil).Once()

	mockBuilder.On("BuildProductDocument", mock.Anything, product, mock.Anything, mock.Anything, mock.Anything).
		Return(domain.ProductDocument{UUID: id}, nil).Once()

	handler := NewSyncHandler(mockRepo, mockBuilder, []string{"en", "th"})

	start := time.Now()
	doc, err := handler.SyncProduct(ctx, id)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, id, doc.UUID)

	// Sequential execution takes ~3×delay ≈ 180ms.
	// Concurrent execution takes ~1×delay ≈ 60ms.
	// Threshold of 2×delay (120ms) distinguishes the two cases unambiguously.
	assert.Less(t, elapsed, 2*delay,
		"SyncProduct must fan out the three DB calls concurrently; elapsed=%s", elapsed)
}
