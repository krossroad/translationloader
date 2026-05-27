// Package dto provides Data Transfer Objects for Elasticsearch document mapping.
package dto

// Label represents a map of locale strings to their corresponding localized labels.
type Label map[string]string

// Brand contains code and label information for a brand.
type Brand struct {
	Code  string `json:"code"`
	Label Label  `json:"label"`
}

// ProductName represents a product name localized for a specific locale.
type ProductName struct {
	Locale string `json:"locale"`
	Data   string `json:"data"`
}

// Property represents a generic property with a code and localized labels.
type Property struct {
	Code  string `json:"code"`
	Label Label  `json:"label"`
}

// ElasticsearchDocument represents the structure of the document stored in Elasticsearch.
type ElasticsearchDocument struct {
	UUID        string            `json:"uuid"`
	SKU         string            `json:"sku"`
	PartNumber  string            `json:"part_number"`
	Brand       Brand             `json:"brand"`
	ProductName []ProductName     `json:"productname"`
	OilGrade    Property          `json:"oil_grade"`
	Attributes  map[string]string `json:"attributes"`
}
