-- name: CreateLink :exec
INSERT INTO links (code, original_url, created_at, expires_at)
VALUES ($1, $2, $3, $4);

-- name: FindLinkByCode :one
SELECT code, original_url, created_at, expires_at
FROM links
WHERE code = $1;
