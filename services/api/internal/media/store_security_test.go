package media

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFileStoreRejectsSymlinkedMediaPaths(t *testing.T) {
	ctx := context.Background()
	root, outside := t.TempDir(), t.TempDir()
	f, err := NewFileStore(root)
	if err != nil { t.Fatal(err) }
	secret := filepath.Join(outside, "private.jpg")
	if err := os.WriteFile(secret, []byte("outside"), 0o600); err != nil { t.Fatal(err) }
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil { t.Skipf("symlink unavailable: %v", err) }
	for _, key := range []string{"linked/private.jpg", "/private.jpg", "../private.jpg", "linked/../other.jpg"} {
		if err := f.Put(ctx, key, "image/jpeg", []byte("overwrite")); err == nil { t.Errorf("Put accepted %q", key) }
		if _, err := f.Get(ctx, key); err == nil { t.Errorf("Get accepted %q", key) }
		if err := f.Delete(ctx, key); err == nil { t.Errorf("Delete accepted %q", key) }
	}
	if data, err := os.ReadFile(secret); err != nil || string(data) != "outside" { t.Fatalf("outside file changed: %q, %v", data, err) }
	if err := os.Symlink(secret, filepath.Join(root, "file.jpg")); err != nil { t.Fatal(err) }
	if _, err := f.Get(ctx, "file.jpg"); err == nil { t.Fatal("read symlinked file") }
	if err := f.Put(ctx, "file.jpg", "image/jpeg", []byte("overwrite")); err == nil { t.Fatal("wrote symlinked file") }
	if err := f.Delete(ctx, "file.jpg"); err == nil { t.Fatal("deleted symlinked file") }
}

func TestFileStoreConcurrentPutKeepsCompleteBlob(t *testing.T) {
	f, err := NewFileStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	ctx := context.Background()
	a, b := bytes.Repeat([]byte("a"), 64<<10), bytes.Repeat([]byte("b"), 64<<10)
	var wg sync.WaitGroup
	for _, data := range [][]byte{a, b} {
		wg.Add(1)
		go func(data []byte) { defer wg.Done(); if err := f.Put(ctx, "same.jpg", "image/jpeg", data); err != nil { t.Errorf("Put: %v", err) } }(data)
	}
	wg.Wait()
	blob, err := f.Get(ctx, "same.jpg")
	if err != nil { t.Fatal(err) }
	if !bytes.Equal(blob.Bytes, a) && !bytes.Equal(blob.Bytes, b) { t.Fatal("concurrent writes produced incomplete blob") }
}
