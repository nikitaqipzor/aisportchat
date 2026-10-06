package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBodyScanMediaQueueRetainsFailureAndSkipsActivePhoto(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	u, err := m.CreateUser(ctx, "body-media-queue@example.com", "hash")
	if err != nil { t.Fatal(err) }
	scan, err := m.CreateBodyScan(ctx, BodyScan{UserID: u.ID})
	if err != nil { t.Fatal(err) }
	if err := m.StageBodyScanPhoto(ctx, u.ID, scan.Scan.ID, "active"); err != nil { t.Fatal(err) }
	if _, err := m.UpsertBodyScanPhoto(ctx, BodyScanPhoto{UserID: u.ID, ScanID: scan.Scan.ID, View: "front", StorageKey: "active"}); err != nil { t.Fatal(err) }
	if err := m.StageBodyScanPhoto(ctx, u.ID, scan.Scan.ID, "orphan"); err != nil { t.Fatal(err) }
	m.mu.Lock()
	m.bodyScanMedia["orphan"] = bodyScanMediaJob{userID: u.ID, readyAt: time.Now().Add(-time.Second)}
	m.bodyScanMedia["active"] = bodyScanMediaJob{userID: u.ID, readyAt: time.Now().Add(-time.Second)}
	m.mu.Unlock()
	removed := map[string]int{}
	if err := m.RetryBodyScanMedia(ctx, u.ID, func(_ context.Context, key string) error { removed[key]++; return errors.New("disk failed") }); err == nil { t.Fatal("expected deletion failure") }
	if removed["active"] != 0 || removed["orphan"] != 1 { t.Fatalf("unsafe cleanup: %+v", removed) }
	if err := m.RetryBodyScanMedia(ctx, u.ID, func(_ context.Context, key string) error { removed[key]++; return nil }); err != nil { t.Fatal(err) }
	if removed["active"] != 0 || removed["orphan"] != 2 { t.Fatalf("retry did not clear orphan: %+v", removed) }
	if len(m.bodyScanMedia) != 0 { t.Fatalf("queue still contains %d jobs", len(m.bodyScanMedia)) }
}

func TestBodyScanMediaQueueExpeditesDeletedAccount(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	u, err := m.CreateUser(ctx, "body-media-deleted@example.com", "hash")
	if err != nil { t.Fatal(err) }
	scan, err := m.CreateBodyScan(ctx, BodyScan{UserID: u.ID})
	if err != nil { t.Fatal(err) }
	if err := m.StageBodyScanPhoto(ctx, u.ID, scan.Scan.ID, "staged"); err != nil { t.Fatal(err) }
	if err := m.DeleteAccount(ctx, u.ID, func(context.Context) error { return nil }); err != nil { t.Fatal(err) }
	removed := false
	if err := m.RetryBodyScanMedia(ctx, "", func(_ context.Context, key string) error { removed = key == "staged"; return nil }); err != nil { t.Fatal(err) }
	if !removed || len(m.bodyScanMedia) != 0 { t.Fatal("deleted account retained staged key") }
}
