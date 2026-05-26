package dto

type Label map[string]string

type Brand struct {
	Code  string `json:"code"`
	Label Label  `json:"label"`
}

type ProductName struct {
	Locale string `json:"locale"`
	Data   string `json:"data"`
}

type Property struct {
	Code  string `json:"code"`
	Label Label  `json:"label"`
}

type ElasticsearchDocument struct {
	UUID        string            `json:"uuid"`
	SKU         string            `json:"sku"`
	PartNumber  string            `json:"part_number"`
	Brand       Brand             `json:"brand"`
	ProductName []ProductName     `json:"productname"`
	OilGrade    Property          `json:"oil_grade"`
	Attributes  map[string]string `json:"attributes"`
}
