package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/rikeshs/translationloader/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestDTOMapping(t *testing.T) {
	t.Run("dbTranslation toDomain", func(t *testing.T) {
		now := time.Now()
		db := dbTranslation{
			ID:         "1",
			EntityType: "product",
			EntityID:   "p1",
			Locale:     "en",
			FieldName:  "name",
			FieldValue: "val",
			UpdatedAt:  now,
		}
		dom := db.toDomain()
		assert.Equal(t, db.ID, dom.ID)
		assert.Equal(t, domain.EntityTypeProduct, dom.EntityType)
		assert.Equal(t, db.EntityID, dom.EntityID)
		assert.Equal(t, db.FieldValue, dom.FieldValue)
	})

	t.Run("dbProduct toDomain", func(t *testing.T) {
		db := dbProduct{
			ID:         "p1",
			SKU:        "S1",
			PartNumber: "PN1",
			Brand:      "B1",
			CategoryID: sql.NullString{String: "C1", Valid: true},
		}
		dom := db.toDomain()
		assert.Equal(t, db.ID, dom.ID)
		assert.Equal(t, "C1", dom.CategoryID)

		db.CategoryID = sql.NullString{String: "", Valid: false}
		dom = db.toDomain()
		assert.Equal(t, "", dom.CategoryID)
	})

	t.Run("dbAttribute toDomain", func(t *testing.T) {
		db := dbAttribute{
			ID:         "a1",
			Code:       "code1",
			MetricUnit: sql.NullString{String: "unit1", Valid: true},
		}
		dom := db.toDomain()
		assert.Equal(t, "unit1", dom.MetricUnit)
	})
}

func TestPostgresProductRepository(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresProductRepository(mock)
	ctx := context.Background()

	t.Run("GetProduct", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "sku", "part_number", "brand", "category_id"}).
			AddRow("p1", "SKU1", "PN1", "B1", "C1")
		mock.ExpectQuery("SELECT id, sku, part_number, brand, category_id FROM product WHERE id = \\$1").
			WithArgs("p1").
			WillReturnRows(rows)

		p, err := repo.GetProduct(ctx, "p1")
		assert.NoError(t, err)
		assert.Equal(t, "p1", p.ID)
	})

	t.Run("GetAttributesByProductID", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "code", "metric_unit"}).
			AddRow("a1", "attr1", "unit1")
		mock.ExpectQuery("SELECT a.id, a.code, a.metric_unit FROM attribute a JOIN product_specification ps ON a.id = ps.attribute_id WHERE ps.product_id = \\$1").
			WithArgs("p1").
			WillReturnRows(rows)

		attrs, err := repo.GetAttributesByProductID(ctx, "p1")
		assert.NoError(t, err)
		assert.Len(t, attrs, 1)
		assert.Equal(t, "attr1", attrs[0].Code)
	})

	t.Run("GetSpecificationsByProductID", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "product_id", "attribute_id", "value"}).
			AddRow("s1", "p1", "a1", "v1")
		mock.ExpectQuery("SELECT id, product_id, attribute_id, value FROM product_specification WHERE product_id = \\$1").
			WithArgs("p1").
			WillReturnRows(rows)

		specs, err := repo.GetSpecificationsByProductID(ctx, "p1")
		assert.NoError(t, err)
		assert.Len(t, specs, 1)
		assert.Equal(t, "v1", specs[0].Value)
	})
}

func TestPostgresProductRepository_GetProduct_NotFound(t *testing.T) {
	// Bug 3: GetProduct wraps pgx.ErrNoRows with fmt.Errorf, so callers cannot use
	// errors.Is to distinguish "not found" from a transient DB error. The fix requires
	// domain.ErrNotFound to be defined and returned (wrapped) when pgx.ErrNoRows occurs.
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresProductRepository(mock)
	ctx := context.Background()

	t.Run("GetProduct returns ErrNotFound when row missing", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, sku, part_number, brand, category_id FROM product WHERE id = \\$1").
			WithArgs("missing-id").
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.GetProduct(ctx, "missing-id")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrNotFound),
			"expected errors.Is(err, domain.ErrNotFound) to be true, got: %v", err)
	})
}

func TestPostgresTranslationLoader(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	loader := NewPostgresTranslationLoader(mock)
	ctx := context.Background()

	t.Run("BulkLoad", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "entity_type", "entity_id", "locale", "field_name", "field_value", "updated_at"}).
			AddRow("t1", "product", "p1", "en", "name", "Name EN", time.Now())
		
		mock.ExpectQuery("SELECT id, entity_type, entity_id, locale, field_name, field_value, updated_at FROM translation WHERE entity_id = ANY\\(\\$1\\) AND locale = ANY\\(\\$2\\)").
			WithArgs([]string{"p1"}, []string{"en"}).
			WillReturnRows(rows)

		res, err := loader.BulkLoad(ctx, []string{"p1"}, []string{"en"})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "Name EN", res["p1"][0].FieldValue)
	})
}
