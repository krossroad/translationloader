package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rikeshs/translationloader/internal/core/domain"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresTranslationLoader struct {
	db DB
}

func NewPostgresTranslationLoader(db DB) *PostgresTranslationLoader {
	return &PostgresTranslationLoader{db: db}
}

func (l *PostgresTranslationLoader) BulkLoad(ctx context.Context, entityIDs []string, locales []string) (map[string][]domain.Translation, error) {
	if len(entityIDs) == 0 {
		return make(map[string][]domain.Translation), nil
	}

	query := `
		SELECT id, entity_type, entity_id, locale, field_name, field_value, updated_at
		FROM translation
		WHERE entity_id = ANY($1)
	`
	args := []interface{}{entityIDs}

	if len(locales) > 0 {
		query += " AND locale = ANY($2)"
		args = append(args, locales)
	}

	rows, err := l.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query translations: %w", err)
	}
	defer rows.Close()

	results := make(map[string][]domain.Translation)
	for rows.Next() {
		var t dbTranslation
		err := rows.Scan(&t.ID, &t.EntityType, &t.EntityID, &t.Locale, &t.FieldName, &t.FieldValue, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}
		results[t.EntityID] = append(results[t.EntityID], t.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}
