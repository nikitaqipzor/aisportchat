package media

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var ErrNotFound = errors.New("media not found")

type Blob struct {
	Bytes    []byte
	MimeType string
}

type Store interface {
	Put(ctx context.Context, key, mimeType string, data []byte) error
	Get(ctx context.Context, key string) (Blob, error)
	Delete(ctx context.Context, key string) error
	DeleteUserMedia(ctx context.Context, userID string) error
}

var canonicalUserID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func userMediaPrefix(userID string) (string, error) {
	if !canonicalUserID.MatchString(userID) { return "", errors.New("invalid user id for media removal") }
	return "body-scans/" + userID, nil
}

type FileStore struct{ Root string }

func NewFileStore(root string) (*FileStore, error) {
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "." || root == "" {
		root = "./data/media"
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	return &FileStore{Root: root}, nil
}

func (f *FileStore) path(key string) (string, error) {
	clean := filepath.Clean("/" + key)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || strings.Contains(clean, "..") {
		return "", errors.New("invalid media key")
	}
	p := filepath.Join(f.Root, clean)
	rootAbs, _ := filepath.Abs(f.Root)
	pAbs, _ := filepath.Abs(p)
	if pAbs != rootAbs && !strings.HasPrefix(pAbs, rootAbs+string(os.PathSeparator)) {
		return "", errors.New("invalid media key")
	}
	return p, nil
}

func (f *FileStore) Put(_ context.Context, key, _ string, data []byte) error {
	p, err := f.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (f *FileStore) Get(_ context.Context, key string) (Blob, error) {
	p, err := f.path(key)
	if err != nil {
		return Blob{}, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Blob{}, ErrNotFound
	}
	if err != nil {
		return Blob{}, err
	}
	mime := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(p), ".png") {
		mime = "image/png"
	}
	return Blob{Bytes: b, MimeType: mime}, nil
}

func (f *FileStore) Delete(_ context.Context, key string) error {
	p, err := f.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); errors.Is(err, os.ErrNotExist) {
		return nil
	} else {
		return err
	}
}

func (f *FileStore) DeleteUserMedia(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil { return err }
	prefix, err := userMediaPrefix(userID)
	if err != nil { return err }
	parent := filepath.Join(f.Root, "body-scans")
	if info, err := os.Lstat(parent); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() { return errors.New("unsafe body-scans directory") }
	} else if errors.Is(err, os.ErrNotExist) { return nil } else { return err }
	p, err := f.path(prefix)
	if err != nil { return err }
	if info, err := os.Lstat(p); err == nil && info.Mode()&os.ModeSymlink != 0 { return errors.New("unsafe user media directory") } else if err != nil && !errors.Is(err, os.ErrNotExist) { return err }
	// RemoveAll does not follow symlinks inside this bounded directory.
	return os.RemoveAll(p)
}

type MemoryStore struct {
	mu    sync.RWMutex
	blobs map[string]Blob
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{blobs: map[string]Blob{}} }
func (m *MemoryStore) Put(_ context.Context, key, mime string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blobs[key] = Blob{Bytes: append([]byte(nil), data...), MimeType: mime}
	return nil
}
func (m *MemoryStore) Get(_ context.Context, key string) (Blob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.blobs[key]
	if !ok {
		return Blob{}, ErrNotFound
	}
	b.Bytes = append([]byte(nil), b.Bytes...)
	return b, nil
}
func (m *MemoryStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.blobs, key)
	return nil
}

func (m *MemoryStore) DeleteUserMedia(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil { return err }
	prefix, err := userMediaPrefix(userID)
	if err != nil { return err }
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.blobs { if strings.HasPrefix(key, prefix+"/") { delete(m.blobs, key) } }
	return nil
}
