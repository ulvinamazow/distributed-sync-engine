package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/domain"
)

type ItemRepository interface {
	Create(ctx context.Context, item *domain.Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	Update(ctx context.Context, item *domain.Item) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.Item, error)
}
