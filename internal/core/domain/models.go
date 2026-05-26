package domain

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type EntityType string

const (
	EntityTypeProduct              EntityType = "product"
	EntityTypeAttribute            EntityType = "attribute"
	EntityTypeProductSpecification EntityType = "product_specification"
)

type Translation struct {
	ID         string
	EntityType EntityType
	EntityID   string
	Locale     string
	FieldName  string
	FieldValue string
	UpdatedAt  time.Time
}

// Translations indexes Translation records by field name.
type Translations map[string]Translation

type Product struct {
	ID         string
	SKU        string
	PartNumber string
	Brand      string
	CategoryID string
}

type Attribute struct {
	ID         string
	Code       string
	MetricUnit string
}

type ProductSpecification struct {
	ID          string
	ProductID   string
	AttributeID string
	Value       string
}

type Label struct {
	En string
	Th string
}

type BrandInfo struct {
	Code  string
	Label Label
}

type ProductName struct {
	Locale string
	Data   string
}

type Property struct {
	Code  string
	Label Label
}

type ProductDocument struct {
	UUID        string
	SKU         string
	PartNumber  string
	Brand       BrandInfo
	ProductName []ProductName
	OilGrade    Property
	Attributes  map[string]string
}
