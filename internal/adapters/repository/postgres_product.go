package repository

import (
	"context"
	"fmt"

	"github.com/rikeshs/translationloader/internal/core/domain"
)

type PostgresProductRepository struct {
	db DB
}

func NewPostgresProductRepository(db DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	var p dbProduct
	err := r.db.QueryRow(ctx, "SELECT id, sku, part_number, brand, category_id FROM product WHERE id = $1", id).Scan(&p.ID, &p.SKU, &p.PartNumber, &p.Brand, &p.CategoryID)
	if err != nil {
		return domain.Product{}, fmt.Errorf("failed to query product: %w", err)
	}
	return p.toDomain(), nil
}

func (r *PostgresProductRepository) GetAttributesByProductID(ctx context.Context, productID string) ([]domain.Attribute, error) {
	rows, err := r.db.Query(ctx, "SELECT a.id, a.code, a.metric_unit FROM attribute a JOIN product_specification ps ON a.id = ps.attribute_id WHERE ps.product_id = $1", productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query attributes: %w", err)
	}
	defer rows.Close()

	var attrs []domain.Attribute
	for rows.Next() {
		var a dbAttribute
		if err := rows.Scan(&a.ID, &a.Code, &a.MetricUnit); err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}
		attrs = append(attrs, a.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during attributes iteration: %w", err)
	}
	return attrs, nil
}

func (r *PostgresProductRepository) GetSpecificationsByProductID(ctx context.Context, productID string) ([]domain.ProductSpecification, error) {
	rows, err := r.db.Query(ctx, "SELECT id, product_id, attribute_id, value FROM product_specification WHERE product_id = $1", productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query specifications: %w", err)
	}
	defer rows.Close()

	var specs []domain.ProductSpecification
	for rows.Next() {
		var s dbSpecification
		if err := rows.Scan(&s.ID, &s.ProductID, &s.AttributeID, &s.Value); err != nil {
			return nil, fmt.Errorf("failed to scan specification: %w", err)
		}
		specs = append(specs, s.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during specifications iteration: %w", err)
	}
	return specs, nil
}
