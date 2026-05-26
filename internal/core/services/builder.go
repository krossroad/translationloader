package services

import (
	"context"
	"fmt"

	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/rikeshs/translationloader/internal/core/ports"
)

type DocumentBuilder struct {
	loader ports.TranslationLoader
}

func NewDocumentBuilder(loader ports.TranslationLoader) *DocumentBuilder {
	return &DocumentBuilder{loader: loader}
}

// BuildProductDocument assembles the Elasticsearch document for a single product.
// It fetches translations for the product, its attributes, and specifications.
func (b *DocumentBuilder) BuildProductDocument(ctx context.Context, p domain.Product, attrs []domain.Attribute, specs []domain.ProductSpecification, locales []string) (domain.ProductDocument, error) {
	entityIDs := b.collectEntityIDs(p, attrs, specs)
	fetchLocales := b.prepareLocales(locales)

	// Bulk load translations in one round-trip
	translations, err := b.loader.BulkLoad(ctx, entityIDs, fetchLocales)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("bulk loading translations: %w", err)
	}

	doc := b.initializeDocument(p, translations)
	b.populateProductNames(&doc, translations, locales, p.ID, p.SKU)
	b.populateAttributes(&doc, attrs, specs, translations)

	return doc, nil
}

func (b *DocumentBuilder) collectEntityIDs(p domain.Product, attrs []domain.Attribute, specs []domain.ProductSpecification) []string {
	ids := []string{p.ID}
	for _, a := range attrs {
		ids = append(ids, a.ID)
	}
	for _, s := range specs {
		ids = append(ids, s.ID)
	}
	return ids
}

func (b *DocumentBuilder) prepareLocales(locales []string) []string {
	hasEn := false
	for _, l := range locales {
		if l == "en" {
			hasEn = true
			break
		}
	}
	if !hasEn {
		return append([]string{"en"}, locales...)
	}
	return locales
}

func (b *DocumentBuilder) initializeDocument(p domain.Product, translations map[string][]domain.Translation) domain.ProductDocument {
	brandLabelEn := b.getTranslation(translations[p.ID], "brand_label", "en")
	if brandLabelEn == "" {
		brandLabelEn = p.Brand
	}
	brandLabelTh := b.getTranslation(translations[p.ID], "brand_label", "th")
	if brandLabelTh == "" {
		brandLabelTh = brandLabelEn
	}

	return domain.ProductDocument{
		UUID:       p.ID,
		SKU:        p.SKU,
		PartNumber: p.PartNumber,
		Brand: domain.BrandInfo{
			Code: p.Brand,
			Label: domain.Label{
				En: brandLabelEn,
				Th: brandLabelTh,
			},
		},
		ProductName: make([]domain.ProductName, 0),
		Attributes:  make(map[string]string),
	}
}

func (b *DocumentBuilder) populateProductNames(doc *domain.ProductDocument, translations map[string][]domain.Translation, locales []string, productID string, sku string) {
	for _, l := range locales {
		name := b.getTranslation(translations[productID], "productname", l)
		if name == "" {
			name = sku
		}
		doc.ProductName = append(doc.ProductName, domain.ProductName{Locale: l, Data: name})
	}
}

func (b *DocumentBuilder) populateAttributes(doc *domain.ProductDocument, attrs []domain.Attribute, specs []domain.ProductSpecification, translations map[string][]domain.Translation) {
	attrMap := make(map[string]domain.Attribute)
	for _, a := range attrs {
		attrMap[a.ID] = a
	}

	for _, s := range specs {
		attr, ok := attrMap[s.AttributeID]
		if !ok {
			continue
		}

		// Use spec value translation if available, otherwise fallback to raw value
		val := b.getTranslation(translations[s.ID], "value_label", "en") // default to en for simplicity here
		if val == "" {
			val = s.Value
		}
		doc.Attributes[attr.Code] = val

		// Special handling for oil_grade as requested in the doc shape
		if attr.Code == "oil_grade" {
			doc.OilGrade = b.mapOilGrade(s, translations[s.ID])
		}
	}
}

func (b *DocumentBuilder) mapOilGrade(spec domain.ProductSpecification, translations []domain.Translation) domain.Property {
	og := domain.Property{
		Code: spec.Value,
		Label: domain.Label{
			En: b.getTranslation(translations, "value_label", "en"),
			Th: b.getTranslation(translations, "value_label", "th"),
		},
	}
	if og.Label.En == "" {
		og.Label.En = spec.Value
	}
	if og.Label.Th == "" {
		og.Label.Th = og.Label.En
	}
	return og
}

func (b *DocumentBuilder) getTranslation(list []domain.Translation, field string, locale string) string {
	// 1. Look for exact locale and field
	for _, t := range list {
		if t.FieldName == field && t.Locale == locale {
			return t.FieldValue
		}
	}
	// 2. Fallback to 'en' if requested was different
	if locale != "en" {
		for _, t := range list {
			if t.FieldName == field && t.Locale == "en" {
				return t.FieldValue
			}
		}
	}
	return ""
}
