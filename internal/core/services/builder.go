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

	doc := b.initializeDocument(p, translations, fetchLocales)
	b.populateProductNames(&doc, translations, fetchLocales, p.ID, p.SKU)
	b.populateAttributes(&doc, attrs, specs, translations, fetchLocales)

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

func (b *DocumentBuilder) initializeDocument(p domain.Product, translations map[string]domain.Translations, fetchLocales []string) domain.ProductDocument {
	productTrans := translations[p.ID]

	brandLabel := make(domain.Label, len(fetchLocales))
	for _, locale := range fetchLocales {
		val := b.getTranslation(productTrans, "brand_label", locale)
		if val == "" {
			val = p.Brand
		}
		brandLabel[locale] = val
	}

	return domain.ProductDocument{
		UUID:       p.ID,
		SKU:        p.SKU,
		PartNumber: p.PartNumber,
		Brand: domain.BrandInfo{
			Code:  p.Brand,
			Label: brandLabel,
		},
		ProductName: make([]domain.ProductName, 0),
		Attributes:  make(map[string]string),
	}
}

func (b *DocumentBuilder) populateProductNames(doc *domain.ProductDocument, translations map[string]domain.Translations, locales []string, productID string, sku string) {
	for _, l := range locales {
		name := b.getTranslation(translations[productID], "productname", l)
		if name == "" {
			name = sku
		}
		doc.ProductName = append(doc.ProductName, domain.ProductName{Locale: l, Data: name})
	}
}

func (b *DocumentBuilder) populateAttributes(doc *domain.ProductDocument, attrs []domain.Attribute, specs []domain.ProductSpecification, translations map[string]domain.Translations, locales []string) {
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
			doc.OilGrade = b.mapOilGrade(s, translations[s.ID], locales)
		}
	}
}

func (b *DocumentBuilder) mapOilGrade(spec domain.ProductSpecification, translations domain.Translations, locales []string) domain.Property {
	og := domain.Property{
		Code:  spec.Value,
		Label: make(domain.Label),
	}
	for _, locale := range locales {
		val := b.getTranslation(translations, "value_label", locale)
		if val == "" {
			val = b.getTranslation(translations, "value_label", "en")
		}
		if val == "" {
			val = spec.Value
		}
		og.Label[locale] = val
	}
	if og.Label["en"] == "" {
		og.Label["en"] = spec.Value
	}
	return og
}

func (b *DocumentBuilder) getTranslation(list domain.Translations, field string, locale string) string {
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
