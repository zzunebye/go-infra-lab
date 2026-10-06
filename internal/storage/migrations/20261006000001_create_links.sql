-- +goose Up
CREATE TABLE links (
    code VARCHAR(16) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE links;
