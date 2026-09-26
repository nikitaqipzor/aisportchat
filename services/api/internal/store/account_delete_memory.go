package store

import (
	"context"
	"strings"
	"errors"
)

func (m *Memory) DeleteAccount(ctx context.Context, userID string, cleanup func(context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return ErrNotFound
	}
	m.pendingMedia[userID] = true
	delete(m.users, userID)
	delete(m.usersByMail, u.Email)
	delete(m.profiles, userID)
	delete(m.goals, userID)
	delete(m.prefs, userID)
	delete(m.onboarding, userID)
	for k, v := range m.sessions {
		if v.UserID == userID { delete(m.sessions, k) }
	}
	for id, v := range m.workouts {
		if v.UserID == userID {
			delete(m.workouts, id)
			delete(m.workoutEx, id)
			delete(m.workoutSets, id)
		}
	}
	records := m.records[:0]
	for _, v := range m.records { if v.UserID != userID { records = append(records, v) } }
	m.records = records
	delete(m.nutrition, userID)
	for id, v := range m.foodItems { if v.OwnerUserID == userID { delete(m.foodItems, id) } }
	delete(m.foodEntries, userID)
	for k := range m.foodOperations { if strings.HasPrefix(k, userID+"\x00") { delete(m.foodOperations, k) } }
	for id, v := range m.recipes { if v.UserID == userID { delete(m.recipes, id) } }
	delete(m.measurements, userID)
	for id, v := range m.bodyScans {
		if v.UserID == userID { delete(m.bodyScans, id); delete(m.bodyScanPhotos, id) }
	}
	for id, v := range m.techniqueAnalyses { if v.UserID == userID { delete(m.techniqueAnalyses, id) } }
	for id, v := range m.recoveryCheckIns { if v.UserID == userID { delete(m.recoveryCheckIns, id) } }
	for id, v := range m.healthSnapshots { if v.UserID == userID { delete(m.healthSnapshots, id) } }
	for id, v := range m.programs { if v.UserID == userID { delete(m.programs, id); delete(m.programSessions, id) } }
	m.mu.Unlock()
	err := cleanup(ctx)
	m.mu.Lock()
	if err != nil { return errors.Join(ErrMediaPending, err) }
	delete(m.pendingMedia, userID)
	return nil
}

func (m *Memory) RetryPendingMedia(ctx context.Context, cleanup func(context.Context, string) error) error {
	m.mu.RLock()
	ids := make([]string, 0, len(m.pendingMedia))
	for id := range m.pendingMedia { ids = append(ids, id) }
	m.mu.RUnlock()
	var all error
	for _, id := range ids {
		if err := cleanup(ctx, id); err != nil { all = errors.Join(all, err); continue }
		m.mu.Lock()
		delete(m.pendingMedia, id)
		m.mu.Unlock()
	}
	return all
}
