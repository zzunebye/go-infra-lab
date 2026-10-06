package link_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-infra-lab/internal/link"
	"go-infra-lab/internal/storage"
)

func TestCreateLinkValidatesAndPersists(t *testing.T) {
	repository := storage.NewMemoryRepository()
	service := link.NewServiceWithGenerator(repository, func() (string, error) { return "aB3kP9x", nil })

	item, err := service.CreateLink(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}
	if item.Code != "aB3kP9x" || item.OriginalURL != "https://example.com/path" {
		t.Fatalf("CreateLink() = %+v", item)
	}
	if item.CreatedAt.IsZero() || item.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt should be UTC, got %v", item.CreatedAt)
	}
	stored, err := repository.FindByCode(context.Background(), item.Code)
	if err != nil || stored != item {
		t.Fatalf("stored link = %+v, error = %v", stored, err)
	}
}

func TestCreateLinkRejectsInvalidURLs(t *testing.T) {
	service := link.NewService(storage.NewMemoryRepository())
	for _, raw := range []string{"example.com", "hello", "ftp://example.com", "https:///missing-host"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := service.CreateLink(context.Background(), raw); !errors.Is(err, link.ErrInvalidURL) {
				t.Fatalf("CreateLink(%q) error = %v, want ErrInvalidURL", raw, err)
			}
		})
	}
}

func TestResolveLinkAndNotFound(t *testing.T) {
	repository := storage.NewMemoryRepository()
	service := link.NewServiceWithGenerator(repository, func() (string, error) { return "abc1234", nil })
	if _, err := service.CreateLink(context.Background(), "https://example.com"); err != nil {
		t.Fatal(err)
	}
	resolved, err := service.ResolveLink(context.Background(), "abc1234")
	if err != nil || resolved != "https://example.com" {
		t.Fatalf("ResolveLink() = %q, %v", resolved, err)
	}
	if _, err := service.ResolveLink(context.Background(), "missing"); !errors.Is(err, link.ErrLinkNotFound) {
		t.Fatalf("ResolveLink() error = %v, want ErrLinkNotFound", err)
	}
}

func TestCreateLinkRetriesCodeCollision(t *testing.T) {
	repository := storage.NewMemoryRepository()
	sequence := []string{"taken01", "fresh01"}
	service := link.NewServiceWithGenerator(repository, func() (string, error) {
		code := sequence[0]
		sequence = sequence[1:]
		return code, nil
	})
	if err := repository.Create(context.Background(), link.Link{Code: "taken01"}); err != nil {
		t.Fatal(err)
	}
	item, err := service.CreateLink(context.Background(), "https://example.com")
	if err != nil || item.Code != "fresh01" {
		t.Fatalf("CreateLink() = %+v, %v", item, err)
	}
}

func TestCreateLinkStopsAfterCollisionLimit(t *testing.T) {
	repository := storage.NewMemoryRepository()
	if err := repository.Create(context.Background(), link.Link{Code: "taken01"}); err != nil {
		t.Fatal(err)
	}
	service := link.NewServiceWithGenerator(repository, func() (string, error) { return "taken01", nil })
	if _, err := service.CreateLink(context.Background(), "https://example.com"); !errors.Is(err, link.ErrCodeGeneration) {
		t.Fatalf("CreateLink() error = %v, want ErrCodeGeneration", err)
	}
}
