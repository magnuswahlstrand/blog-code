package middleware

import "context"

type db interface {
	get(ctx context.Context, idempotencyKey string) (RequestRecording, bool, error)
	update(ctx context.Context, idempotencyKey string, record RequestRecording) error
}
