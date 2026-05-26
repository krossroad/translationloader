package repository

import (
	"database/sql"
	"time"

	"github.com/rikeshs/translationloader/internal/core/domain"
)

type dbTranslation struct {
	ID         string
	EntityType string
	EntityID   string
	Locale     string
	FieldName  string
	FieldValue string
	UpdatedAt  time.Time
}

func (t dbTranslation) toDomain() domain.Translation {
	return domain.Translation{
		ID:         t.ID,
		EntityType: domain.EntityType(t.EntityType),
		EntityID:   t.EntityID,
		Locale:     t.Locale,
		FieldName:  t.FieldName,
		FieldValue: t.FieldValue,
		UpdatedAt:  t.UpdatedAt,
	}
}

type dbProduct struct {
	ID         string
	SKU        string
	PartNumber string
	Brand      string
	CategoryID sql.NullString
}

func (p dbProduct) toDomain() domain.Product {
	return domain.Product{
		ID:         p.ID,
		SKU:        p.SKU,
		PartNumber: p.PartNumber,
		Brand:      p.Brand,
		CategoryID: p.CategoryID.String,
	}
}

type dbAttribute struct {
	ID         string
	Code       string
	MetricUnit sql.NullString
}

func (a dbAttribute) toDomain() domain.Attribute {
	return domain.Attribute{
		ID:         a.ID,
		Code:       a.Code,
		MetricUnit: a.MetricUnit.String,
	}
}

type dbSpecification struct {
	ID          string
	ProductID   string
	AttributeID string
	Value       string
}

func (s dbSpecification) toDomain() domain.ProductSpecification {
	return domain.ProductSpecification{
		ID:          s.ID,
		ProductID:   s.ProductID,
		AttributeID: s.AttributeID,
		Value:       s.Value,
	}
}
