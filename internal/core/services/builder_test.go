package services

import (
	"context"
	"testing"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDocumentBuilder_BuildProductDocument(t *testing.T) {
	ctx := context.Background()
	product := domain.Product{ID: "p-1", SKU: "SKU-1", Brand: "bosch"}
	attrs := []domain.Attribute{{ID: "a-1", Code: "oil_grade"}}
	specs := []domain.ProductSpecification{{ID: "s-1", ProductID: "p-1", AttributeID: "a-1", Value: "5w30"}}
	locales := []string{"en", "th"}

	tests := []struct {
		name          string
		setupMocks    func(m *mocks.TranslationLoader)
		expectedCheck func(t *testing.T, doc domain.ProductDocument, err error)
	}{
		{
			name: "success with complete translations",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string][]domain.Translation{
					"p-1": {
						{EntityID: "p-1", Locale: "en", FieldName: "productname", FieldValue: "Oil 1L"},
						{EntityID: "p-1", Locale: "th", FieldName: "productname", FieldValue: "น้ำมัน 1 ลิตร"},
					},
					"s-1": {
						{EntityID: "s-1", Locale: "en", FieldName: "value_label", FieldValue: "5W-30"},
						{EntityID: "s-1", Locale: "th", FieldName: "value_label", FieldValue: "5W-30 TH"},
					},
				}, nil)
			},
			expectedCheck: func(t *testing.T, doc domain.ProductDocument, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "p-1", doc.UUID)
				assert.Equal(t, "Oil 1L", doc.ProductName[0].Data)
				assert.Equal(t, "น้ำมัน 1 ลิตร", doc.ProductName[1].Data)
				assert.Equal(t, "5W-30", doc.Attributes["oil_grade"])
				assert.Equal(t, "5W-30", doc.OilGrade.Label.En)
				assert.Equal(t, "5W-30 TH", doc.OilGrade.Label.Th)
			},
		},
		{
			name: "fallback to en when th is missing",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string][]domain.Translation{
					"p-1": {
						{EntityID: "p-1", Locale: "en", FieldName: "productname", FieldValue: "Oil 1L"},
					},
					"s-1": {
						{EntityID: "s-1", Locale: "en", FieldName: "value_label", FieldValue: "5W-30"},
					},
				}, nil)
			},
			expectedCheck: func(t *testing.T, doc domain.ProductDocument, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "Oil 1L", doc.ProductName[0].Data)
				assert.Equal(t, "Oil 1L", doc.ProductName[1].Data) // Fallback
				assert.Equal(t, "5W-30", doc.OilGrade.Label.En)
				assert.Equal(t, "5W-30", doc.OilGrade.Label.Th) // Fallback
			},
		},
		{
			name: "bulk load error",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string][]domain.Translation{}, assert.AnError)
			},
			expectedCheck: func(t *testing.T, doc domain.ProductDocument, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), assert.AnError.Error())
			},
		},
		{
			name: "no translations found - use default values",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string][]domain.Translation{}, nil)
			},
			expectedCheck: func(t *testing.T, doc domain.ProductDocument, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "SKU-1", doc.ProductName[0].Data) // Defaults to SKU if no name
				assert.Equal(t, "5w30", doc.Attributes["oil_grade"]) // Defaults to spec value if no label
				assert.Equal(t, "5w30", doc.OilGrade.Label.En)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := mocks.NewTranslationLoader(t)
			builder := NewDocumentBuilder(mockLoader)
			tt.setupMocks(mockLoader)

			doc, err := builder.BuildProductDocument(ctx, product, attrs, specs, locales)
			tt.expectedCheck(t, doc, err)
		})
	}
}
