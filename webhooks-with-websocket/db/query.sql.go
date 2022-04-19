// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: query.sql

package webhooksdb

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const insertWebhook = `-- name: InsertWebhook :one
INSERT INTO webhooks (subscription_id, payload)
VALUES ($1, $2)
RETURNING id
`

type InsertWebhookParams struct {
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	Payload        json.RawMessage `json:"payload"`
}

func (q *Queries) InsertWebhook(ctx context.Context, arg InsertWebhookParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, insertWebhook, arg.SubscriptionID, arg.Payload)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const listWebhooks = `-- name: ListWebhooks :many
SELECT id, subscription_id, created_at, acked_at, payload
FROM webhooks
WHERE subscription_id = $2
  AND acked_at IS NULL
ORDER BY created_at
LIMIT $1
`

type ListWebhooksParams struct {
	Limit          int32     `json:"limit"`
	SubscriptionID uuid.UUID `json:"subscription_id"`
}

func (q *Queries) ListWebhooks(ctx context.Context, arg ListWebhooksParams) ([]Webhook, error) {
	rows, err := q.db.QueryContext(ctx, listWebhooks, arg.Limit, arg.SubscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Webhook
	for rows.Next() {
		var i Webhook
		if err := rows.Scan(
			&i.ID,
			&i.SubscriptionID,
			&i.CreatedAt,
			&i.AckedAt,
			&i.Payload,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setPublished = `-- name: SetPublished :execrows
UPDATE webhooks
SET acked_at = NOW()
WHERE id = $1
  AND subscription_id = $2
`

type SetPublishedParams struct {
	ID             uuid.UUID `json:"id"`
	SubscriptionID uuid.UUID `json:"subscription_id"`
}

func (q *Queries) SetPublished(ctx context.Context, arg SetPublishedParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, setPublished, arg.ID, arg.SubscriptionID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
