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

func TestPostgresBodyScanUpsertChecksOwnerAndDraft(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is not set")
	}
	pg, err := NewPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Close()
	ctx := context.Background()
	owner, err := pg.CreateUser(ctx, fmt.Sprintf("scan-owner-%d@example.com", time.Now().UnixNano()), "hash")
	if err != nil {
		t.Fatal(err)
	}
	other, err := pg.CreateUser(ctx, fmt.Sprintf("scan-other-%d@example.com", time.Now().UnixNano()), "hash")
	if err != nil {
		t.Fatal(err)
	}
	scan, err := pg.CreateBodyScan(ctx, BodyScan{UserID: owner.ID})
	if err != nil {
		t.Fatal(err)
	}
	defer pg.DeleteBodyScan(ctx, owner.ID, scan.Scan.ID)
	photo := BodyScanPhoto{ScanID: scan.Scan.ID, UserID: other.ID, View: "front", StorageKey: "foreign-key", MimeType: "image/jpeg", QualityStatus: "accepted"}
	if _, err := pg.UpsertBodyScanPhoto(ctx, photo); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign owner inserted photo: %v", err)
	}
	photo.UserID = owner.ID
	for _, view := range []string{"front", "side", "back"} {
		photo.View = view
		photo.StorageKey = view + "-original"
		if err := pg.StageBodyScanPhoto(ctx, owner.ID, scan.Scan.ID, photo.StorageKey); err != nil { t.Fatal(err) }
		if _, err := pg.UpsertBodyScanPhoto(ctx, photo); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pg.CompleteBodyScan(ctx, owner.ID, scan.Scan.ID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	photo.View, photo.StorageKey = "front", "replacement-after-complete"
	if _, err := pg.UpsertBodyScanPhoto(ctx, photo); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("completed scan accepted replacement: %v", err)
	}
	current, err := pg.GetBodyScan(ctx, owner.ID, scan.Scan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Scan.Status != "completed" || len(current.Photos) != 3 || current.Photos[0].StorageKey != "front-original" {
		t.Fatalf("unexpected stored scan after rejected writes: %+v", current)
	}
}

func TestPostgresBodyScanCleanupSurvivesReplacementAndDeletion(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" { t.Skip("POSTGRES_TEST_DSN is not set") }
	pg, err := NewPostgres(dsn)
	if err != nil { t.Fatal(err) }
	defer pg.Close()
	ctx := context.Background()
	owner, err := pg.CreateUser(ctx, fmt.Sprintf("scan-cleanup-%d@example.com", time.Now().UnixNano()), "hash")
	if err != nil { t.Fatal(err) }
	scan, err := pg.CreateBodyScan(ctx, BodyScan{UserID: owner.ID})
	if err != nil { t.Fatal(err) }
	defer pg.DeleteBodyScan(ctx, owner.ID, scan.Scan.ID)
	photo := BodyScanPhoto{ScanID: scan.Scan.ID, UserID: owner.ID, View: "front", MimeType: "image/jpeg", QualityStatus: "accepted"}
	for _, key := range []string{"first", "second"} {
		photo.StorageKey = fmt.Sprintf("body-scans/%s/%s/%s", owner.ID, scan.Scan.ID, key)
		if err := pg.StageBodyScanPhoto(ctx, owner.ID, scan.Scan.ID, photo.StorageKey); err != nil { t.Fatal(err) }
		if _, err := pg.UpsertBodyScanPhoto(ctx, photo); err != nil { t.Fatal(err) }
	}
	removed := []string{}
	fail := errors.New("disk unavailable")
	if err := pg.RetryBodyScanMedia(ctx, owner.ID, func(_ context.Context, key string) error { return fail }); !errors.Is(err, fail) { t.Fatalf("expected retryable failure: %v", err) }
	if err := pg.RetryBodyScanMedia(ctx, owner.ID, func(_ context.Context, key string) error { removed = append(removed, key); return nil }); err != nil { t.Fatal(err) }
	if len(removed) != 1 || removed[0] != fmt.Sprintf("body-scans/%s/%s/first", owner.ID, scan.Scan.ID) { t.Fatalf("replacement cleanup: %v", removed) }
	if err := pg.DeleteBodyScan(ctx, owner.ID, scan.Scan.ID); err != nil { t.Fatal(err) }
	if err := pg.RetryBodyScanMedia(ctx, owner.ID, func(_ context.Context, key string) error { removed = append(removed, key); return nil }); err != nil { t.Fatal(err) }
	if len(removed) != 2 || removed[1] != fmt.Sprintf("body-scans/%s/%s/second", owner.ID, scan.Scan.ID) { t.Fatalf("deletion cleanup: %v", removed) }
}
