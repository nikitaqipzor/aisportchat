package media

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteUserMediaBoundedAndOrphans(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	f, err := NewFileStore(root)
	if err != nil { t.Fatal(err) }
	u1 := "12345678-1234-1234-1234-123456789abc"
	u2 := "12345678-1234-1234-1234-123456789abd"
	for _, id := range []string{u1,u2} {
		if err := f.Put(ctx, "body-scans/"+id+"/old/orphan.jpg", "image/jpeg", []byte(id)); err != nil { t.Fatal(err) }
	}
	if err := f.DeleteUserMedia(ctx, u1); err != nil { t.Fatal(err) }
	if _, err := f.Get(ctx, "body-scans/"+u1+"/old/orphan.jpg"); !errors.Is(err, ErrNotFound) { t.Fatalf("first user's photo still exists: %v", err) }
	if _, err := f.Get(ctx, "body-scans/"+u2+"/old/orphan.jpg"); err != nil { t.Fatalf("second user's photo removed: %v", err) }
	if err := f.DeleteUserMedia(ctx, "../"+u2); err == nil { t.Fatal("unsafe ID accepted") }
}

func TestDeleteUserMediaRejectsSymlinkedParent(t *testing.T) {
	ctx := context.Background()
	root, outside := t.TempDir(), t.TempDir()
	f, err := NewFileStore(root)
	if err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(outside, "private.jpg"), []byte("other"), 0o600); err != nil { t.Fatal(err) }
	if err := os.Symlink(outside, filepath.Join(root, "body-scans")); err != nil { t.Skipf("symlink unavailable: %v", err) }
	if err := f.DeleteUserMedia(ctx, "12345678-1234-1234-1234-123456789abc"); err == nil { t.Fatal("symlinked parent accepted") }
	if _, err := os.Stat(filepath.Join(outside, "private.jpg")); err != nil { t.Fatalf("outside file touched: %v", err) }
}
