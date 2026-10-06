//go:build cgo

package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/media"
)

func TestPostgresAccountDeletionCascadeAndDurableMediaRetry(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" { t.Skip("POSTGRES_TEST_DSN is not set; run migrations including 000021 before testing") }
	p, err := NewPostgres(dsn)
	if err != nil { t.Fatal(err) }
	defer p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	owner, err := p.CreateUser(ctx, fmt.Sprintf("delete-owner-%d@example.com", time.Now().UnixNano()), "hash")
	if err != nil { t.Fatal(err) }
	other, err := p.CreateUser(ctx, fmt.Sprintf("delete-other-%d@example.com", time.Now().UnixNano()), "hash")
	if err != nil { t.Fatal(err) }
	defer func() { _, _ = p.query(context.Background(), `DELETE FROM users WHERE id=$1::uuid`, sp(other.ID)) }()
	insert := func(sql string, params ...*string) {
		t.Helper()
		if _, err := p.query(ctx, sql, params...); err != nil { t.Fatal(err) }
	}
	insert(`INSERT INTO user_profiles(user_id,injuries) VALUES($1::uuid,'["shoulder"]'::jsonb)`, sp(owner.ID))
	insert(`INSERT INTO workouts(user_id,status,environment) VALUES($1::uuid,'planned','gym')`, sp(owner.ID))
	insert(`INSERT INTO health_daily_snapshots(user_id,local_date,source_package,captured_at) VALUES($1::uuid,current_date,'test',now())`, sp(owner.ID))
	insert(`INSERT INTO food_items(id,name,kcal_per_100g,owner_user_id) VALUES($1,'custom',100,$2::uuid)`, sp("custom-delete-"+owner.ID), sp(owner.ID))
	insert(`INSERT INTO body_scans(id,user_id) VALUES(gen_random_uuid(),$1::uuid)`, sp(owner.ID))
	insert(`INSERT INTO user_profiles(user_id) VALUES($1::uuid)`, sp(other.ID))
	blobs, err := media.NewFileStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	ownerKey := "body-scans/"+owner.ID+"/orphan.jpg"
	otherKey := "body-scans/"+other.ID+"/orphan.jpg"
	if err := blobs.Put(ctx, ownerKey, "image/jpeg", []byte("owner")); err != nil { t.Fatal(err) }
	if err := blobs.Put(ctx, otherKey, "image/jpeg", []byte("other")); err != nil { t.Fatal(err) }
	mediaError := errors.New("simulated media outage")
	err = p.DeleteAccount(ctx, owner.ID, func(context.Context) error { cancel(); return mediaError })
	if !errors.Is(err, ErrMediaPending) { t.Fatalf("want durable pending state, got %v", err) }
	check := func(query string, want int) {
		t.Helper()
		rows, err := p.query(context.Background(), query, sp(owner.ID))
		if err != nil { t.Fatal(err) }
		if got := *rows[0][0]; got != fmt.Sprint(want) { t.Fatalf("%s: got %s want %d", query, got, want) }
	}
	for _, table := range []string{"users", "user_profiles", "workouts", "health_daily_snapshots", "food_items", "body_scans"} {
		column := "user_id"
		if table == "users" { column = "id" }
		if table == "food_items" { column = "owner_user_id" }
		check("SELECT count(*)::text FROM "+table+" WHERE "+column+"=$1::uuid", 0)
	}
	check(`SELECT count(*)::text FROM pending_media_deletions WHERE user_id=$1::uuid`, 1)
	rows, err := p.query(context.Background(), `SELECT count(*)::text FROM user_profiles WHERE user_id=$1::uuid`, sp(other.ID))
	if err != nil || *rows[0][0] != "1" { t.Fatalf("second user profile damaged: %v %v", rows, err) }
	called := 0
	if err := p.RetryPendingMedia(context.Background(), func(retryCtx context.Context, userID string) error {
		if userID == owner.ID { called++ }; return blobs.DeleteUserMedia(retryCtx, userID)
	}); err != nil { t.Fatal(err) }
	if called == 0 { t.Fatal("durable job not retried after request cancellation") }
	check(`SELECT count(*)::text FROM pending_media_deletions WHERE user_id=$1::uuid`, 0)
	if _, err := blobs.Get(context.Background(), ownerKey); !errors.Is(err, media.ErrNotFound) { t.Fatalf("owner blob remains: %v", err) }
	if _, err := blobs.Get(context.Background(), otherKey); err != nil { t.Fatalf("other blob removed: %v", err) }
}
