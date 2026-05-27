package services

import (
	"context"
	"testing"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/tests/mocks"
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
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string]domain.Translations{
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
				assert.Equal(t, "5W-30", doc.OilGrade.Label["en"])
				assert.Equal(t, "5W-30 TH", doc.OilGrade.Label["th"])
			},
		},
		{
			name: "fallback to en when th is missing",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string]domain.Translations{
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
				assert.Equal(t, "5W-30", doc.OilGrade.Label["en"])
				assert.Equal(t, "5W-30", doc.OilGrade.Label["th"]) // Fallback
			},
		},
		{
			name: "bulk load error",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string]domain.Translations{}, assert.AnError)
			},
			expectedCheck: func(t *testing.T, doc domain.ProductDocument, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), assert.AnError.Error())
			},
		},
		{
			name: "no translations found - use default values",
			setupMocks: func(m *mocks.TranslationLoader) {
				m.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).Return(map[string]domain.Translations{}, nil)
			},
			expectedCheck: func(t *testing.T, doc domain.ProductDocument, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "SKU-1", doc.ProductName[0].Data)    // Defaults to SKU if no name
				assert.Equal(t, "5w30", doc.Attributes["oil_grade"]) // Defaults to spec value if no label
				assert.Equal(t, "5w30", doc.OilGrade.Label["en"])
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

// TestDocumentBuilder_ThirdLocale_InLabel verifies that a locale beyond "en" and "th"
// (e.g. "de") is correctly stored in domain.Label after BuildProductDocument.
// This test currently fails to compile because domain.Label is a struct with fixed
// En and Th fields rather than a map[string]string.
func TestDocumentBuilder_ThirdLocale_InLabel(t *testing.T) {
	ctx := context.Background()
	product := domain.Product{ID: "p-5", SKU: "SKU-5", Brand: "castrol"}
	attrs := []domain.Attribute{{ID: "a-5", Code: "oil_grade"}}
	specs := []domain.ProductSpecification{{ID: "s-5", ProductID: "p-5", AttributeID: "a-5", Value: "5w30"}}
	locales := []string{"en", "th", "de"}

	mockLoader := mocks.NewTranslationLoader(t)
	mockLoader.On("BulkLoad", mock.Anything, mock.Anything, mock.Anything).
		Return(map[string]domain.Translations{
			"p-5": {
				{EntityID: "p-5", Locale: "en", FieldName: "productname", FieldValue: "Engine Oil"},
				{EntityID: "p-5", Locale: "th", FieldName: "productname", FieldValue: "น้ำมันเครื่อง"},
				{EntityID: "p-5", Locale: "de", FieldName: "productname", FieldValue: "Motoröl"},
			},
			"s-5": {
				{EntityID: "s-5", Locale: "en", FieldName: "value_label", FieldValue: "5W-30"},
				{EntityID: "s-5", Locale: "th", FieldName: "value_label", FieldValue: "5W-30 TH"},
				{EntityID: "s-5", Locale: "de", FieldName: "value_label", FieldValue: "5W-30 DE"},
			},
		}, nil).Once()

	builder := NewDocumentBuilder(mockLoader)
	doc, err := builder.BuildProductDocument(ctx, product, attrs, specs, locales)

	assert.NoError(t, err)
	// domain.Label must be map[string]string after the fix; this line fails to compile
	// on the current struct-based Label.
	assert.Equal(t, "5W-30 DE", doc.OilGrade.Label["de"],
		"OilGrade.Label must carry the German translation when 'de' is a requested locale")
}

func TestDocumentBuilder_BuildProductDocument_EnLocaleInjectedIntoProductName(t *testing.T) {
	// Bug 1: populateProductNames is called with the original (un-enriched) locales slice,
	// so the English product name that was fetched via fetchLocales is never written into doc.ProductName.
	ctx := context.Background()
	product := domain.Product{ID: "p-2", SKU: "SKU-2", Brand: "castrol"}
	attrs := []domain.Attribute{}
	specs := []domain.ProductSpecification{}

	// Caller requests only "th" — prepareLocales should inject "en" into the BulkLoad call.
	requestedLocales := []string{"th"}

	mockLoader := mocks.NewTranslationLoader(t)

	// BulkLoad will be called with ["en", "th"] because prepareLocales injects "en".
	mockLoader.On("BulkLoad", mock.Anything, []string{"p-2"}, []string{"en", "th"}).
		Return(map[string]domain.Translations{
			"p-2": {
				{EntityID: "p-2", Locale: "en", FieldName: "productname", FieldValue: "Engine Oil"},
				{EntityID: "p-2", Locale: "th", FieldName: "productname", FieldValue: "น้ำมันเครื่อง"},
			},
		}, nil).Once()

	builder := NewDocumentBuilder(mockLoader)
	doc, err := builder.BuildProductDocument(ctx, product, attrs, specs, requestedLocales)

	assert.NoError(t, err)
	// populateProductNames must use the enriched fetchLocales (["en","th"]), not the original
	// requestedLocales (["th"]). So doc.ProductName must contain both locales.
	assert.Len(t, doc.ProductName, 2, "expected both 'en' and 'th' entries in ProductName")

	localeSet := make(map[string]string)
	for _, pn := range doc.ProductName {
		localeSet[pn.Locale] = pn.Data
	}
	assert.Equal(t, "Engine Oil", localeSet["en"], "English product name must be populated")
	assert.Equal(t, "น้ำมันเครื่อง", localeSet["th"], "Thai product name must be populated")
}
