package link

import "time"

type Link struct {
	Code        string     `json:"code"`
	OriginalURL string     `json:"originalUrl"`
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}
