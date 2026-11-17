package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/domain"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/repository"
)

type ItemService struct {
	repo repository.ItemRepository
}

func NewItemService(repo repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) Create(ctx context.Context, name, description string) (*domain.Item, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name can not be empty")
	}

	item := &domain.Item{
		ID:          uuid.New(),
		Name:        name,
		Description: strings.TrimSpace(description),
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ItemService) Get(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ItemService) Update(ctx context.Context, id uuid.UUID, name, description string) (*domain.Item, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("ad boş ola bilməz")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = name
	existing.Description = strings.TrimSpace(description)

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *ItemService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ItemService) List(ctx context.Context) ([]*domain.Item, error) {
	return s.repo.List(ctx)
}
