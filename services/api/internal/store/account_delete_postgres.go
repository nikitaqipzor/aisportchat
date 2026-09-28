//go:build cgo

package store

import (
	"context"
	"errors"
	"time"
)

func (p *Postgres) DeleteAccount(ctx context.Context, userID string, cleanup func(context.Context) error) error {
	err := p.withTx(ctx, func() error {
		rows, err := p.queryLocked(ctx, `SELECT id::text FROM users WHERE id=$1::uuid FOR UPDATE`, sp(userID))
		if err != nil { return err }
		if len(rows) == 0 { return ErrNotFound }
		if _, err = p.queryLocked(ctx, `INSERT INTO pending_media_deletions(user_id) VALUES($1::uuid) ON CONFLICT DO NOTHING`, sp(userID)); err != nil { return err }
		_, err = p.queryLocked(ctx, `DELETE FROM users WHERE id=$1::uuid`, sp(userID))
		return err
	})
	if err != nil { return err }
	// The deletion job is durable before file cleanup begins. Interrupted HTTP
	// requests cannot cancel this attempt; the worker retries any failure.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := cleanup(cleanupCtx); err != nil { return errors.Join(ErrMediaPending, err) }
	if _, err := p.query(cleanupCtx, `DELETE FROM pending_media_deletions WHERE user_id=$1::uuid`, sp(userID)); err != nil { return errors.Join(ErrMediaPending, err) }
	return nil
}

func (p *Postgres) RetryPendingMedia(ctx context.Context, cleanup func(context.Context, string) error) error {
	rows, err := p.query(ctx, `SELECT user_id::text FROM pending_media_deletions ORDER BY created_at LIMIT 100`)
	if err != nil { return err }
	var all error
	for _, row := range rows {
		id := *row[0]
		if err := cleanup(ctx, id); err != nil { all = errors.Join(all, err); continue }
		if _, err := p.query(ctx, `DELETE FROM pending_media_deletions WHERE user_id=$1::uuid`, sp(id)); err != nil { all = errors.Join(all, err) }
	}
	return all
}
