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
	if err := rejectSymlink(root); err != nil { return nil, err }
	return &FileStore{Root: root}, nil
}

func (f *FileStore) path(key string) (string, error) {
	if key == "" || filepath.IsAbs(key) {
		return "", errors.New("invalid media key")
	}
	for _, part := range strings.Split(key, string(os.PathSeparator)) {
		if part == "" || part == "." || part == ".." { return "", errors.New("invalid media key") }
	}
	return filepath.Join(f.Root, key), nil
}

// Refuse symlinked directories and files so a stored key cannot reach outside
// the private media root. The media directory must be owned by the API process.
func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil { return err }
	if info.Mode()&os.ModeSymlink != 0 { return errors.New("symlink in media path") }
	return nil
}

func (f *FileStore) checkPath(p string, create bool) error {
	if err := rejectSymlink(f.Root); err != nil { return err }
	rel, err := filepath.Rel(f.Root, p)
	if err != nil { return err }
	current := f.Root
	parts := strings.Split(rel, string(os.PathSeparator))
	for _, part := range parts[:len(parts)-1] {
		current = filepath.Join(current, part)
		if create {
			if err := os.Mkdir(current, 0o700); err != nil && !errors.Is(err, os.ErrExist) { return err }
		}
		if err := rejectSymlink(current); err != nil { return err }
		info, err := os.Stat(current)
		if err != nil { return err }
		if !info.IsDir() { return errors.New("non-directory in media path") }
	}
	if err := rejectSymlink(p); err != nil && !errors.Is(err, os.ErrNotExist) { return err }
	return nil
}

func (f *FileStore) Put(_ context.Context, key, _ string, data []byte) error {
	p, err := f.path(key)
	if err != nil {
		return err
	}
	if err := f.checkPath(p, true); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(p), ".media-*")
	if err != nil { return err }
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil { _ = file.Close(); return err }
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), p)
}

func (f *FileStore) Get(_ context.Context, key string) (Blob, error) {
	p, err := f.path(key)
	if err != nil {
		return Blob{}, err
	}
	if err := f.checkPath(p, false); errors.Is(err, os.ErrNotExist) { return Blob{}, ErrNotFound } else if err != nil { return Blob{}, err }
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
	if err := f.checkPath(p, false); errors.Is(err, os.ErrNotExist) { return nil } else if err != nil { return err }
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
	if err := rejectSymlink(f.Root); err != nil { return err }
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
