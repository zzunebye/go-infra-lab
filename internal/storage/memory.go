package storage

import (
	"context"
	"sync"

	"go-infra-lab/internal/link"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	links map[string]link.Link
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{links: make(map[string]link.Link)}
}

func (r *MemoryRepository) Create(ctx context.Context, item link.Link) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.links[item.Code]; exists {
		return link.ErrCodeCollision
	}
	r.links[item.Code] = item
	return nil
}

func (r *MemoryRepository) FindByCode(ctx context.Context, code string) (link.Link, error) {
	if err := ctx.Err(); err != nil {
		return link.Link{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.links[code]
	if !exists {
		return link.Link{}, link.ErrLinkNotFound
	}
	return item, nil
}
