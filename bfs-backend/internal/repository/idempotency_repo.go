package repository

import (
	"context"
	"encoding/json"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/idempotency"
)

type IdempotencyRepository interface {
	Get(ctx context.Context, scope, key string) (*ent.Idempotency, error)
	SaveIfAbsent(ctx context.Context, scope, key string, response map[string]any, ttl time.Duration) (*ent.Idempotency, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

type idempotencyRepo struct {
	client *ent.Client
}

func NewIdempotencyRepository(client *ent.Client) IdempotencyRepository {
	return &idempotencyRepo{client: client}
}

func (r *idempotencyRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *idempotencyRepo) Get(ctx context.Context, scope, key string) (*ent.Idempotency, error) {
	e, err := r.ec(ctx).Idempotency.Query().
		Where(
			idempotency.ScopeEQ(scope),
			idempotency.KeyEQ(key),
			idempotency.ExpiresAtGT(time.Now()),
		).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *idempotencyRepo) SaveIfAbsent(ctx context.Context, scope, key string, response map[string]any, ttl time.Duration) (*ent.Idempotency, error) {
	// Marshal response to JSON
	var responseJSON []byte
	var err error
	if response != nil {
		responseJSON, err = json.Marshal(response)
		if err != nil {
			return nil, err
		}
	}

	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	expiresAt := time.Now().Add(ttl)

	// Try to create; if a constraint error occurs, the record already exists.
	created, err := r.ec(ctx).Idempotency.Create().
		SetScope(scope).
		SetKey(key).
		SetResponse(responseJSON).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			// Already exists, return existing record
			return r.Get(ctx, scope, key)
		}
		return nil, translateError(err)
	}

	return created, nil
}

func (r *idempotencyRepo) CleanupExpired(ctx context.Context) (int64, error) {
	n, err := r.ec(ctx).Idempotency.Delete().
		Where(idempotency.ExpiresAtLT(time.Now())).
		Exec(ctx)
	if err != nil {
		return 0, translateError(err)
	}
	return int64(n), nil
}

// GetResponseMap unmarshals the JSON response bytes from an ent.Idempotency into a map.
func GetResponseMap(record *ent.Idempotency) (map[string]any, error) {
	if record == nil || len(record.Response) == 0 {
		return nil, nil
	}
	var response map[string]any
	if err := json.Unmarshal(record.Response, &response); err != nil {
		return nil, err
	}
	return response, nil
}
