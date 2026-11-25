package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/domain"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/repository"
)

type ItemRepo struct {
	pool *pgxpool.Pool
}

func NewItemRepo(pool *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{pool: pool}
}

var _ repository.ItemRepository = (*ItemRepo)(nil)

func (r *ItemRepo) Create(ctx context.Context, item *domain.Item) error {
	query := `INSERT INTO items (id, name, description, created_at, updated_at)
	VALUES ($1, $2, $3, NOW(), NOW()) 
	RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		item.ID, item.Name, item.Description).Scan(&item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("item not created: %w", err)
	}
	return nil
}

func (r *ItemRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	var item domain.Item
	query := `SELECT id, name, description, created_at, updated_at 
	FROM items 
	WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.Name, &item.Description,
		&item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("item not read: %w", err)
	}
	return &item, nil
}

func (r *ItemRepo) Update(ctx context.Context, item *domain.Item) error {
	query := `UPDATE items SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		item.Name, item.Description, item.ID,
	).Scan(&item.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("item not updated: %w", err)
	}
	return nil
}

func (r *ItemRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM items WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("item not deleted: %w", err)
	}
	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *ItemRepo) List(ctx context.Context) ([]*domain.Item, error) {
	query := `SELECT id, name, description, created_at, updated_at 
	FROM items 
	ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("items not listed: %w", err)
	}

	defer rows.Close()

	var items []*domain.Item
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("row not readed: %w", err)
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while reading lines: %w", err)
	}
	return items, nil
}

func (r *ItemRepo) UpsertFromRemote(ctx context.Context, item *domain.Item) error {
	query := ` INSERT INTO items (id, name, description, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (ID) DO UPDATE SET
	name = EXCLUDED.name,
	description = EXCLUDED.description,
	updated_at = EXCLUDED.updated_at
	WHERE items.updated_at < EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query, item.ID, item.Name,
		item.Description, item.CreatedAt, item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("upsert unsuccesful: %w", err)
	}

	return nil
}
