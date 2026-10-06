package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-infra-lab/internal/link"
	"go-infra-lab/internal/storage/db"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresRepository(ctx context.Context, databaseURL string) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create database connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &PostgresRepository{pool: pool, queries: db.New(pool)}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, item link.Link) error {
	expiresAt := pgtype.Timestamptz{}
	if item.ExpiresAt != nil {
		expiresAt = pgtype.Timestamptz{Time: *item.ExpiresAt, Valid: true}
	}
	err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		Code:        item.Code,
		OriginalUrl: item.OriginalURL,
		CreatedAt:   pgtype.Timestamptz{Time: item.CreatedAt, Valid: true},
		ExpiresAt:   expiresAt,
	})
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return fmt.Errorf("create link: %w", link.ErrCodeCollision)
	}
	return fmt.Errorf("create link: %w", err)
}

func (r *PostgresRepository) FindByCode(ctx context.Context, code string) (link.Link, error) {
	row, err := r.queries.FindLinkByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return link.Link{}, link.ErrLinkNotFound
		}
		return link.Link{}, fmt.Errorf("find link by code: %w", err)
	}
	item := link.Link{
		Code:        row.Code,
		OriginalURL: row.OriginalUrl,
		CreatedAt:   row.CreatedAt.Time,
	}
	if row.ExpiresAt.Valid {
		expiresAt := row.ExpiresAt.Time
		item.ExpiresAt = &expiresAt
	}
	return item, nil
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}
