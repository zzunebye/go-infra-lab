package link

import (
	"context"
	"errors"
)

var (
	ErrInvalidURL     = errors.New("invalid URL")
	ErrLinkNotFound   = errors.New("link not found")
	ErrCodeCollision  = errors.New("short code already exists")
	ErrCodeGeneration = errors.New("could not generate a unique short code")
)

type Repository interface {
	Create(ctx context.Context, link Link) error
	FindByCode(ctx context.Context, code string) (Link, error)
}
