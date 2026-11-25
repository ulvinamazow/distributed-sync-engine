package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/domain"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/repository"
)

type Service struct {
	repo   repository.ItemRepository
	peers  []string
	client *http.Client
	logger *slog.Logger
}

func New(repo repository.ItemRepository, peers []string,
	logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		peers:  peers,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

type PullRequest struct {
	Peer  string `json:"peer"`
	Items int    `json:"items"`
	Error string `json:"error,omitempty"`
}

func (s *Service) PullFromPeers(ctx context.Context) []PullRequest {
	var results []PullRequest
	for _, peer := range s.peers {
		n, err := s.pullFromPeer(ctx, peer)
		res := PullRequest{Peer: peer, Items: n}

		if err != nil {
			res.Error = err.Error()
			s.logger.Warn("peer sync unsuccesful", "peer", peer, "err", err)
		} else {
			s.logger.Info("peer sync successful", "peer", peer, "items", n)
		}

		results = append(results, res)
	}
	return results
}

func (s *Service) pullFromPeer(ctx context.Context, peer string) (int, error) {
	url := peer + "/internal/items"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GET %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("peer %s status %d returned", peer, resp.StatusCode)
	}

	var dtos []itemDTO
	if err := json.NewDecoder(resp.Body).Decode(&dtos); err != nil {
		return 0, fmt.Errorf("response parse : %w", err)
	}

	count := 0
	for _, dto := range dtos {
		it, err := dto.toDomain()
		if err != nil {
			s.logger.Warn("item parse unsuccessful", "err", err, "id", dto.ID)
			continue
		}

		if err := s.repo.UpsertFromRemote(ctx, it); err != nil {
			s.logger.Warn("upsert unsuccessful", "err", err, "id", it.ID)
			continue
		}
		count++
	}
	return count, nil
}

type itemDTO struct {
	ID          string `json:"ID"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (d itemDTO) toDomain() (*domain.Item, error) {
	// uuid.Parse — string-i UUID-yə çevirir, format səhvdirsə xəta qaytarır.
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return nil, fmt.Errorf("uuid parse: %w", err)
	}
	// time.RFC3339Nano — "2006-01-02T15:04:05.999999999Z07:00"
	// Niyə Nano? — DB-də microsaniyə dəqiqliyi var; RFC3339 (saniyə) kifayət etmir
	createdAt, err := time.Parse(time.RFC3339Nano, d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("created_at parse: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updated_at parse: %w", err)
	}
	return &domain.Item{
		ID: id, Name: d.Name, Description: d.Description,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}
