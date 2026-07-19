package database

import (
	"context"
	"errors"
	"fmt"

	"backend/src/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnnouncementRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewAnnouncementRepositoryPGX(pool *pgxpool.Pool) *AnnouncementRepositoryPGX {
	return &AnnouncementRepositoryPGX{pool: pool}
}

func (r *AnnouncementRepositoryPGX) Create(ctx context.Context, announcement *domain.Announcement) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_announcement($1, $2, $3, $4)",
		announcement.Title, announcement.Content, announcement.ImagePath, announcement.PublisherID,
	).Scan(&announcement.ID, &announcement.CreatedAt, &announcement.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating announcement: %w", err)
	}
	return nil
}

func (r *AnnouncementRepositoryPGX) GetByID(ctx context.Context, id string) (*domain.Announcement, error) {
	var a domain.Announcement
	err := r.pool.QueryRow(ctx, "SELECT * FROM get_announcement_by_id($1)", id,
	).Scan(&a.ID, &a.Title, &a.Content, &a.ImagePath, &a.PublisherID, &a.PublisherName, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAnnouncementNotFound
		}
		return nil, fmt.Errorf("getting announcement by ID: %w", err)
	}
	return &a, nil
}

func (r *AnnouncementRepositoryPGX) List(ctx context.Context, limit, offset int) ([]domain.Announcement, int, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_announcements($1, $2)", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing announcements: %w", err)
	}

	announcements, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Announcement, error) {
		var a domain.Announcement
		err := row.Scan(&a.ID, &a.Title, &a.Content, &a.ImagePath, &a.PublisherID, &a.PublisherName, &a.CreatedAt, &a.UpdatedAt)
		return a, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listing announcements: %w", err)
	}

	if announcements == nil {
		announcements = []domain.Announcement{}
	}

	var total int
	err = r.pool.QueryRow(ctx, "SELECT * FROM count_announcements()").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting announcements: %w", err)
	}

	return announcements, total, nil
}

func (r *AnnouncementRepositoryPGX) Update(ctx context.Context, announcement *domain.Announcement) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_announcement($1, $2, $3, $4)",
		announcement.ID, announcement.Title, announcement.Content, announcement.ImagePath,
	).Scan(&announcement.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrAnnouncementNotFound
		}
		return fmt.Errorf("updating announcement: %w", err)
	}
	return nil
}

func (r *AnnouncementRepositoryPGX) Delete(ctx context.Context, id string) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_announcement($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting announcement: %w", err)
	}
	if !deleted {
		return domain.ErrAnnouncementNotFound
	}
	return nil
}
