-- name: InsertWebhook :one
INSERT INTO webhooks (subscription_id, payload)
VALUES ($1, $2)
RETURNING id;

-- name: ListWebhooks :many
SELECT *
FROM webhooks
WHERE subscription_id = $2
  AND acked_at IS NULL
ORDER BY created_at
LIMIT $1;


-- name: SetPublished :execrows
UPDATE webhooks
SET acked_at = NOW()
WHERE id = $1
  AND subscription_id = $2;
