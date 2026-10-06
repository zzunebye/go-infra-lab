package link

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"
)

const (
	codeLength     = 7
	maxCreateTries = 5
	base62         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type CodeGenerator func() (string, error)

type LinkService struct {
	repository Repository
	generate   CodeGenerator
	now        func() time.Time
}

// creating LinkService -> Sort of static function?
func NewService(repository Repository) *LinkService {
	return NewServiceWithGenerator(repository, func() (string, error) {
		return generateCode(rand.Reader)
	})
}

func NewServiceWithGenerator(repository Repository, generator CodeGenerator) *LinkService {
	return &LinkService{repository: repository, generate: generator, now: time.Now}
}

func (s *LinkService) CreateLink(ctx context.Context, originalURL string) (Link, error) {
	if !isValidURL(originalURL) {
		return Link{}, ErrInvalidURL
	}

	for range maxCreateTries {
		code, err := s.generate()
		if err != nil {
			return Link{}, errors.Join(ErrCodeGeneration, err)
		}
		item := Link{Code: code, OriginalURL: originalURL, CreatedAt: s.now().UTC()}
		if err := s.repository.Create(ctx, item); err != nil {
			if errors.Is(err, ErrCodeCollision) {
				continue
			}
			return Link{}, err
		}
		return item, nil
	}
	return Link{}, ErrCodeGeneration
}

func (s *LinkService) GetLink(ctx context.Context, code string) (Link, error) {
	return s.repository.FindByCode(ctx, code)
}

func (s *LinkService) ResolveLink(ctx context.Context, code string) (string, error) {
	item, err := s.GetLink(ctx, code)
	if err != nil {
		return "", err
	}
	return item.OriginalURL, nil
}

func isValidURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return scheme == "http" || scheme == "https"
}

func generateCode(source io.Reader) (string, error) {
	const alphabetSize = len(base62)
	const limit = 256 - (256 % alphabetSize)

	code := make([]byte, codeLength)
	buffer := make([]byte, codeLength*2)
	for position := 0; position < len(code); {
		if _, err := io.ReadFull(source, buffer); err != nil {
			return "", err
		}
		for _, value := range buffer {
			if int(value) >= limit {
				continue
			}
			code[position] = base62[int(value)%alphabetSize]
			position++
			if position == len(code) {
				break
			}
		}
	}
	return string(code), nil
}
