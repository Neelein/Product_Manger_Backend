package database

import (
	"context"
	"sync"
	"time"

	"backend/src/domain"

	"github.com/google/uuid"
)

type InMemoryRepository struct {
	mu       sync.RWMutex
	products map[string]domain.Product
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		products: map[string]domain.Product{},
	}
}

func (r *InMemoryRepository) Create(
	ctx context.Context,
	product *domain.Product,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	product.ID = uuid.New().String()
	product.CreatedAt = now
	product.UpdatedAt = now

	r.products[product.ID] = *product
	return nil
}

func (r *InMemoryRepository) List(
	ctx context.Context,
) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Product, 0, len(r.products))
	for _, p := range r.products {
		result = append(result, p)
	}
	return result, nil
}

func (r *InMemoryRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return &p, nil
}

func (r *InMemoryRepository) Update(
	ctx context.Context,
	product *domain.Product,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.products[product.ID]
	if !ok {
		return domain.ErrProductNotFound
	}

	product.CreatedAt = existing.CreatedAt
	product.UpdatedAt = time.Now()
	r.products[product.ID] = *product
	return nil
}

func (r *InMemoryRepository) Delete(
	ctx context.Context,
	id string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.products[id]; !ok {
		return domain.ErrProductNotFound
	}

	delete(r.products, id)
	return nil
}
