package storage_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"go-infra-lab/internal/link"
	"go-infra-lab/internal/storage"
)

func TestMemoryRepositoryConcurrentCreateAndFind(t *testing.T) {
	const count = 100
	repository := storage.NewMemoryRepository()
	var group sync.WaitGroup
	errCh := make(chan error, count)
	for index := range count {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			code := fmt.Sprintf("code%03d", index)
			if err := repository.Create(context.Background(), link.Link{Code: code}); err != nil {
				errCh <- err
				return
			}
			if _, err := repository.FindByCode(context.Background(), code); err != nil {
				errCh <- err
			}
		}(index)
	}
	group.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}
