//go:build cgo

package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestPostgresRefreshSessionConsumedOnce(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" { t.Skip("POSTGRES_TEST_DSN is not set") }
	pg, err := NewPostgres(dsn)
	if err != nil { t.Fatal(err) }
	defer pg.Close()
	ctx := context.Background()
	user, err := pg.CreateUser(ctx, fmt.Sprintf("refresh-%d@example.com", time.Now().UnixNano()), "hash")
	if err != nil { t.Fatal(err) }
	defer pg.DeleteAccount(ctx, user.ID, func(context.Context) error { return nil })
	hash := fmt.Sprintf("refresh-session-%d", time.Now().UnixNano())
	if err := pg.SaveRefreshSession(ctx, RefreshSession{TokenHash: hash, UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour)}); err != nil { t.Fatal(err) }
	session, err := pg.ConsumeRefreshSession(ctx, hash)
	if err != nil { t.Fatal(err) }
	if session.UserID != user.ID || session.RevokedAt == nil { t.Fatalf("unexpected consumed session: %+v", session) }
	if _, err := pg.ConsumeRefreshSession(ctx, hash); !errors.Is(err, ErrNotFound) { t.Fatalf("reused consumed session: %v", err) }
	expiredHash := hash + "-expired"
	if err := pg.SaveRefreshSession(ctx, RefreshSession{TokenHash: expiredHash, UserID: user.ID, ExpiresAt: time.Now().Add(-time.Second)}); err != nil { t.Fatal(err) }
	if _, err := pg.ConsumeRefreshSession(ctx, expiredHash); !errors.Is(err, ErrNotFound) { t.Fatalf("consumed expired session: %v", err) }
}
