package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type bodyScanMediaJob struct {
	userID string
	readyAt time.Time
}

func (m *Memory) StageBodyScanPhoto(_ context.Context, userID, scanID, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	scan, ok := m.bodyScans[scanID]
	if !ok || scan.UserID != userID { return ErrNotFound }
	if scan.Status != "draft" { return ErrInvalidState }
	m.bodyScanMedia[key] = bodyScanMediaJob{userID: userID, readyAt: time.Now().Add(time.Hour)}
	return nil
}

func (m *Memory) RetryBodyScanMedia(ctx context.Context, userID string, remove func(context.Context, string) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all error
	for key, job := range m.bodyScanMedia {
		_, userExists := m.users[job.userID]
		if userID != "" && job.userID != userID || userExists && job.readyAt.After(time.Now()) { continue }
		active := false
		for _, photos := range m.bodyScanPhotos {
			for _, photo := range photos { if photo.StorageKey == key { active = true; break } }
			if active { break }
		}
		if !active {
			if err := remove(ctx, key); err != nil {
				all = errors.Join(all, fmt.Errorf("delete body scan media %s: %w", key, err))
				continue
			}
		}
		delete(m.bodyScanMedia, key)
	}
	return all
}
