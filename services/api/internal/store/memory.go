package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type Memory struct {
	mu                sync.RWMutex
	users             map[string]User
	usersByMail       map[string]string
	profiles          map[string]Profile
	goals             map[string]Goal
	prefs             map[string]TrainingPreferences
	onboarding        map[string]bool
	sessions          map[string]RefreshSession
	workouts          map[string]Workout
	workoutEx         map[string][]WorkoutExercise
	workoutSets       map[string][]WorkoutSet
	records           []PersonalRecord
	nutrition         map[string]NutritionProfile
	foodItems         map[string]FoodItem
	foodEntries       map[string][]FoodEntry
	recipes           map[string]Recipe
	measurements      map[string][]BodyMeasurement
	bodyScans         map[string]BodyScan
	bodyScanPhotos    map[string]map[string]BodyScanPhoto
	techniqueAnalyses map[string]TechniqueAnalysis
	recoveryCheckIns  map[string]RecoveryCheckIn
	healthSnapshots   map[string]HealthDailySnapshot
	programs          map[string]Program
	programSessions   map[string][]ProgramSession
}

func NewMemory() *Memory {
	return &Memory{
		users:             map[string]User{},
		usersByMail:       map[string]string{},
		profiles:          map[string]Profile{},
		goals:             map[string]Goal{},
		prefs:             map[string]TrainingPreferences{},
		onboarding:        map[string]bool{},
		sessions:          map[string]RefreshSession{},
		workouts:          map[string]Workout{},
		workoutEx:         map[string][]WorkoutExercise{},
		workoutSets:       map[string][]WorkoutSet{},
		records:           []PersonalRecord{},
		nutrition:         map[string]NutritionProfile{},
		foodItems:         seedFoodItems(),
		foodEntries:       map[string][]FoodEntry{},
		recipes:           map[string]Recipe{},
		measurements:      map[string][]BodyMeasurement{},
		bodyScans:         map[string]BodyScan{},
		bodyScanPhotos:    map[string]map[string]BodyScanPhoto{},
		techniqueAnalyses: map[string]TechniqueAnalysis{},
		recoveryCheckIns:  map[string]RecoveryCheckIn{},
		healthSnapshots:   map[string]HealthDailySnapshot{},
		programs:          map[string]Program{},
		programSessions:   map[string][]ProgramSession{},
	}
}

func (m *Memory) CreateUser(_ context.Context, email, passwordHash string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	email = strings.ToLower(strings.TrimSpace(email))
	if _, ok := m.usersByMail[email]; ok {
		return User{}, ErrEmailExists
	}
	id := newID()
	u := User{ID: id, Email: email, PasswordHash: passwordHash, Status: "active", CreatedAt: time.Now().UTC()}
	m.users[id] = u
	m.usersByMail[email] = id
	return u, nil
}

func (m *Memory) FindUserByEmail(_ context.Context, email string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.usersByMail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, ErrNotFound
	}
	return m.users[id], nil
}

func (m *Memory) FindUserByID(_ context.Context, id string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (m *Memory) UpsertProfile(_ context.Context, p Profile) (Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.profiles[p.UserID]
	if p.BirthDate != nil {
		current.BirthDate = p.BirthDate
	}
	if p.Gender != nil {
		current.Gender = p.Gender
	}
	if p.HeightCM != nil {
		current.HeightCM = p.HeightCM
	}
	if p.WeightKG != nil {
		current.WeightKG = p.WeightKG
	}
	if p.ExperienceLevel != nil {
		current.ExperienceLevel = p.ExperienceLevel
	}
	if p.AgeYears != nil {
		current.AgeYears = p.AgeYears
	}
	if p.Injuries != nil {
		current.Injuries = append([]string(nil), p.Injuries...)
	}
	if p.Limitations != nil {
		current.Limitations = append([]string(nil), p.Limitations...)
	}
	if p.UnitSystem != "" {
		current.UnitSystem = p.UnitSystem
	}
	if current.UnitSystem == "" {
		current.UnitSystem = "metric"
	}
	current.UserID = p.UserID
	current.UpdatedAt = time.Now().UTC()
	m.profiles[p.UserID] = current
	return current, nil
}

func (m *Memory) GetProfile(_ context.Context, userID string) (Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.profiles[userID]
	if !ok {
		return Profile{UserID: userID, UnitSystem: "metric", Injuries: []string{}, Limitations: []string{}}, nil
	}
	p.Injuries = append([]string(nil), p.Injuries...)
	p.Limitations = append([]string(nil), p.Limitations...)
	return p, nil
}

func (m *Memory) SetGoal(_ context.Context, g Goal) (Goal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g.StartedAt = time.Now().UTC()
	m.goals[g.UserID] = g
	return g, nil
}

func (m *Memory) GetGoal(_ context.Context, userID string) (Goal, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	g, ok := m.goals[userID]
	if !ok {
		return Goal{}, ErrNotFound
	}
	return g, nil
}

func (m *Memory) SetTrainingPreferences(_ context.Context, p TrainingPreferences) (TrainingPreferences, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	p.Environments = append([]string(nil), p.Environments...)
	p.EquipmentIDs = append([]string(nil), p.EquipmentIDs...)
	m.prefs[p.UserID] = p
	return p, nil
}

func (m *Memory) GetTrainingPreferences(_ context.Context, userID string) (TrainingPreferences, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.prefs[userID]
	if !ok {
		return TrainingPreferences{}, ErrNotFound
	}
	p.Environments = append([]string(nil), p.Environments...)
	p.EquipmentIDs = append([]string(nil), p.EquipmentIDs...)
	return p, nil
}

func (m *Memory) SaveAthleteProfile(_ context.Context, userID string, update AthleteProfileUpdate) (AthleteProfileUpdate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile := update.Profile
	profile.UserID = userID
	profile.Injuries = append([]string(nil), profile.Injuries...)
	profile.Limitations = append([]string(nil), profile.Limitations...)
	if profile.UnitSystem == "" {
		profile.UnitSystem = "metric"
	}
	profile.UpdatedAt = time.Now().UTC()

	goal := update.Goal
	goal.UserID = userID
	goal.StartedAt = time.Now().UTC()

	prefs := update.TrainingPreferences
	prefs.UserID = userID
	prefs.Environments = append([]string(nil), prefs.Environments...)
	prefs.EquipmentIDs = append([]string(nil), prefs.EquipmentIDs...)
	prefs.UpdatedAt = time.Now().UTC()

	// One critical section makes the three map replacements visible atomically.
	m.profiles[userID] = profile
	m.goals[userID] = goal
	m.prefs[userID] = prefs
	return AthleteProfileUpdate{Profile: profile, Goal: goal, TrainingPreferences: prefs}, nil
}

func (m *Memory) GetOnboardingStatus(_ context.Context, userID string) (OnboardingStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, pOK := m.profiles[userID]
	_, gOK := m.goals[userID]
	t, tOK := m.prefs[userID]
	profileOK := pOK && p.HeightCM != nil && p.WeightKG != nil && p.ExperienceLevel != nil
	trainingOK := tOK && len(t.Environments) > 0 && t.WorkoutsPerWeek > 0 && t.SessionMinutes > 0
	return OnboardingStatus{
		ProfileCompleted:  profileOK,
		GoalCompleted:     gOK,
		TrainingCompleted: trainingOK,
		Completed:         m.onboarding[userID] && profileOK && gOK && trainingOK,
	}, nil
}

func (m *Memory) CompleteOnboarding(ctx context.Context, userID string) (OnboardingStatus, error) {
	status, _ := m.GetOnboardingStatus(ctx, userID)
	if !status.ProfileCompleted || !status.GoalCompleted || !status.TrainingCompleted {
		return status, ErrInvalidState
	}
	m.mu.Lock()
	m.onboarding[userID] = true
	m.mu.Unlock()
	status.Completed = true
	return status, nil
}

func (m *Memory) SaveRefreshSession(_ context.Context, s RefreshSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *Memory) GetRefreshSession(_ context.Context, tokenHash string) (RefreshSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[tokenHash]
	if !ok {
		return RefreshSession{}, ErrNotFound
	}
	return s, nil
}

func (m *Memory) RevokeRefreshSession(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[tokenHash]
	if !ok {
		return ErrNotFound
	}
	now := time.Now().UTC()
	s.RevokedAt = &now
	m.sessions[tokenHash] = s
	return nil
}

func (m *Memory) RevokeAllUserSessions(_ context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	for k, s := range m.sessions {
		if s.UserID == userID && s.RevokedAt == nil {
			s.RevokedAt = &now
			m.sessions[k] = s
		}
	}
	return nil
}

func (m *Memory) CreateWorkout(_ context.Context, w Workout, exercises []WorkoutExercise) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w.ID == "" {
		w.ID = newID()
	}
	if w.Status == "" {
		w.Status = "planned"
	}
	w.CreatedAt = time.Now().UTC()
	for i := range exercises {
		if exercises[i].ID == "" {
			exercises[i].ID = newID()
		}
		exercises[i].WorkoutID = w.ID
		exercises[i].Position = i + 1
	}
	m.workouts[w.ID] = w
	m.workoutEx[w.ID] = cloneWorkoutExercises(exercises)
	m.workoutSets[w.ID] = nil
	return m.detailsLocked(w.ID), nil
}

func (m *Memory) GetWorkout(_ context.Context, userID, workoutID string) (WorkoutDetails, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) StartWorkout(_ context.Context, userID, workoutID string) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	if w.Status != "planned" {
		return WorkoutDetails{}, ErrInvalidState
	}
	now := time.Now().UTC()
	w.Status = "active"
	w.StartedAt = &now
	m.workouts[workoutID] = w
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) UpsertWorkoutSet(_ context.Context, userID, workoutID string, set WorkoutSet) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	if w.Status != "active" {
		return WorkoutDetails{}, ErrInvalidState
	}
	validExercise := false
	for _, ex := range m.workoutEx[workoutID] {
		if ex.ID == set.WorkoutExerciseID {
			validExercise = true
			if set.SetNumber < 1 || set.SetNumber > ex.TargetSets+3 {
				return WorkoutDetails{}, ErrInvalidState
			}
			break
		}
	}
	if !validExercise {
		return WorkoutDetails{}, ErrNotFound
	}
	if set.ID == "" {
		set.ID = newID()
	}
	now := time.Now().UTC()
	set.CompletedAt = &now
	sets := m.workoutSets[workoutID]
	replaced := false
	for i := range sets {
		if sets[i].WorkoutExerciseID == set.WorkoutExerciseID && sets[i].SetNumber == set.SetNumber {
			set.ID = sets[i].ID
			sets[i] = set
			replaced = true
			break
		}
	}
	if !replaced {
		sets = append(sets, set)
	}
	m.workoutSets[workoutID] = sets
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) ReplaceWorkoutExercise(_ context.Context, userID, workoutID, workoutExerciseID string, replacement WorkoutExercise) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	if w.Status == "completed" || w.Status == "cancelled" {
		return WorkoutDetails{}, ErrInvalidState
	}
	exercises := m.workoutEx[workoutID]
	idx := -1
	for i := range exercises {
		if exercises[i].ID == workoutExerciseID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return WorkoutDetails{}, ErrNotFound
	}
	for _, s := range m.workoutSets[workoutID] {
		if s.WorkoutExerciseID == workoutExerciseID {
			return WorkoutDetails{}, ErrInvalidState
		}
	}
	replacement.ID = workoutExerciseID
	replacement.WorkoutID = workoutID
	replacement.Position = exercises[idx].Position
	exercises[idx] = replacement
	m.workoutEx[workoutID] = exercises
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) CompleteWorkout(_ context.Context, userID, workoutID string) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	if w.Status != "active" {
		return WorkoutDetails{}, ErrInvalidState
	}
	now := time.Now().UTC()
	w.Status = "completed"
	w.CompletedAt = &now
	if w.StartedAt != nil {
		seconds := int(now.Sub(*w.StartedAt).Seconds())
		if seconds < 0 {
			seconds = 0
		}
		w.DurationSeconds = &seconds
	}
	var volume float64
	completedByExercise := map[string]int{}
	for _, s := range m.workoutSets[workoutID] {
		if s.Weight != nil && s.Repetitions > 0 {
			volume += *s.Weight * float64(s.Repetitions)
		}
		completedByExercise[s.WorkoutExerciseID]++
	}
	var targetSets, completedSets int
	for _, ex := range m.workoutEx[workoutID] {
		targetSets += ex.TargetSets
		done := completedByExercise[ex.ID]
		if done > ex.TargetSets {
			done = ex.TargetSets
		}
		completedSets += done
	}
	w.TotalVolume = volume
	if targetSets > 0 {
		w.CompletionPercent = math.Round((float64(completedSets)/float64(targetSets))*10000) / 100
	}
	w.EndedEarly = completedSets < targetSets
	m.workouts[workoutID] = w
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) CancelWorkout(_ context.Context, userID, workoutID string) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	if w.Status != "planned" && w.Status != "active" {
		return WorkoutDetails{}, ErrInvalidState
	}
	now := time.Now().UTC()
	w.Status = "cancelled"
	w.CompletedAt = nil
	if w.StartedAt != nil {
		seconds := int(now.Sub(*w.StartedAt).Seconds())
		if seconds < 0 {
			seconds = 0
		}
		w.DurationSeconds = &seconds
	}
	m.workouts[workoutID] = w
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) ListWorkouts(_ context.Context, userID string, limit int) ([]WorkoutDetails, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 500 {
		limit = 20
	}
	items := make([]Workout, 0)
	for _, w := range m.workouts {
		if w.UserID == userID {
			items = append(items, w)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]WorkoutDetails, 0, len(items))
	for _, w := range items {
		out = append(out, m.detailsLocked(w.ID))
	}
	return out, nil
}

func (m *Memory) LastCompletedWorkout(_ context.Context, userID, muscle, environment string) (WorkoutDetails, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var best *Workout
	for _, w := range m.workouts {
		if w.UserID != userID || w.Status != "completed" || w.Muscle != muscle || w.Environment != environment {
			continue
		}
		if best == nil || w.CreatedAt.After(best.CreatedAt) {
			copy := w
			best = &copy
		}
	}
	if best == nil {
		return WorkoutDetails{}, ErrNotFound
	}
	return m.detailsLocked(best.ID), nil
}

func (m *Memory) LastExercisePerformance(_ context.Context, userID, exerciseID string) (ExercisePerformance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest *Workout
	var workoutExerciseIDs map[string]bool
	for workoutID, w := range m.workouts {
		if w.UserID != userID || w.Status != "completed" {
			continue
		}
		ids := map[string]bool{}
		for _, ex := range m.workoutEx[workoutID] {
			if ex.ExerciseID == exerciseID {
				ids[ex.ID] = true
			}
		}
		if len(ids) == 0 {
			continue
		}
		when := w.CreatedAt
		if w.CompletedAt != nil {
			when = *w.CompletedAt
		}
		if latest == nil {
			copy := w
			latest = &copy
			workoutExerciseIDs = ids
			continue
		}
		latestWhen := latest.CreatedAt
		if latest.CompletedAt != nil {
			latestWhen = *latest.CompletedAt
		}
		if when.After(latestWhen) {
			copy := w
			latest = &copy
			workoutExerciseIDs = ids
		}
	}
	if latest == nil {
		return ExercisePerformance{}, ErrNotFound
	}

	var best *WorkoutSet
	for _, set := range m.workoutSets[latest.ID] {
		if !workoutExerciseIDs[set.WorkoutExerciseID] || set.CompletedAt == nil {
			continue
		}
		if best == nil || valueFloat(set.Weight) > valueFloat(best.Weight) || (valueFloat(set.Weight) == valueFloat(best.Weight) && set.Repetitions > best.Repetitions) {
			copy := set
			best = &copy
		}
	}
	if best == nil || best.CompletedAt == nil {
		return ExercisePerformance{}, ErrNotFound
	}
	return ExercisePerformance{
		ExerciseID: exerciseID, Weight: cloneFloat(best.Weight), Repetitions: best.Repetitions,
		RPE: cloneFloat(best.RPE), RIR: cloneFloat(best.RIR), CompletedAt: *best.CompletedAt,
	}, nil
}

func (m *Memory) ActiveWorkout(_ context.Context, userID string) (WorkoutDetails, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var best *Workout
	for _, w := range m.workouts {
		if w.UserID != userID || w.Status != "active" {
			continue
		}
		if best == nil || w.CreatedAt.After(best.CreatedAt) {
			copy := w
			best = &copy
		}
	}
	if best == nil {
		return WorkoutDetails{}, ErrNotFound
	}
	return m.detailsLocked(best.ID), nil
}

func (m *Memory) SetWorkoutFavorite(_ context.Context, userID, workoutID string, favorite bool) (WorkoutDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[workoutID]
	if !ok {
		return WorkoutDetails{}, ErrNotFound
	}
	if w.UserID != userID {
		return WorkoutDetails{}, ErrForbidden
	}
	w.Favorite = favorite
	m.workouts[workoutID] = w
	return m.detailsLocked(workoutID), nil
}

func (m *Memory) BestPersonalRecord(_ context.Context, userID, exerciseID, recordType string) (PersonalRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var best *PersonalRecord
	for _, record := range m.records {
		if record.UserID != userID || record.ExerciseID != exerciseID || record.RecordType != recordType {
			continue
		}
		if best == nil || record.Value > best.Value || (record.Value == best.Value && record.AchievedAt.After(best.AchievedAt)) {
			copy := record
			best = &copy
		}
	}
	if best == nil {
		return PersonalRecord{}, ErrNotFound
	}
	return clonePersonalRecord(*best), nil
}

func (m *Memory) SavePersonalRecord(_ context.Context, in PersonalRecord) (PersonalRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.workouts[in.WorkoutID]
	if !ok {
		return PersonalRecord{}, ErrNotFound
	}
	if w.UserID != in.UserID {
		return PersonalRecord{}, ErrForbidden
	}
	if in.ID == "" {
		in.ID = newID()
	}
	if in.AchievedAt.IsZero() {
		in.AchievedAt = time.Now().UTC()
	}
	m.records = append(m.records, clonePersonalRecord(in))
	return clonePersonalRecord(in), nil
}

func (m *Memory) ListWorkoutPersonalRecords(_ context.Context, userID, workoutID string) ([]PersonalRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if w, ok := m.workouts[workoutID]; !ok {
		return nil, ErrNotFound
	} else if w.UserID != userID {
		return nil, ErrForbidden
	}
	out := make([]PersonalRecord, 0)
	for _, record := range m.records {
		if record.UserID == userID && record.WorkoutID == workoutID {
			out = append(out, clonePersonalRecord(record))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AchievedAt.After(out[j].AchievedAt) })
	return out, nil
}

func (m *Memory) ListPersonalRecords(_ context.Context, userID string, limit int) ([]PersonalRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := make([]PersonalRecord, 0)
	for _, record := range m.records {
		if record.UserID == userID {
			out = append(out, clonePersonalRecord(record))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AchievedAt.After(out[j].AchievedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *Memory) UpsertNutritionProfile(_ context.Context, in NutritionProfile) (NutritionProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	in.UpdatedAt = time.Now().UTC()
	m.nutrition[in.UserID] = in
	return in, nil
}

func (m *Memory) GetNutritionProfile(_ context.Context, userID string) (NutritionProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out, ok := m.nutrition[userID]
	if !ok {
		return NutritionProfile{}, ErrNotFound
	}
	return out, nil
}

func (m *Memory) SearchFoodItems(_ context.Context, userID, query string, limit int) ([]FoodItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]FoodItem, 0)
	for _, item := range m.foodItems {
		if item.OwnerUserID != "" && item.OwnerUserID != userID {
			continue
		}
		haystack := strings.ToLower(item.Name + " " + item.Brand)
		if q == "" || strings.Contains(haystack, q) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *Memory) GetFoodItem(_ context.Context, userID, foodID string) (FoodItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out, ok := m.foodItems[foodID]
	if !ok || (out.OwnerUserID != "" && out.OwnerUserID != userID) {
		return FoodItem{}, ErrNotFound
	}
	return out, nil
}

func (m *Memory) FindFoodByBarcode(_ context.Context, userID, barcode string) (FoodItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	barcode = strings.TrimSpace(barcode)
	for _, item := range m.foodItems {
		if item.Barcode == barcode && (item.OwnerUserID == "" || item.OwnerUserID == userID) {
			return item, nil
		}
	}
	return FoodItem{}, ErrNotFound
}

func (m *Memory) CreateCustomFood(_ context.Context, userID string, food FoodItem) (FoodItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[userID]; !ok {
		return FoodItem{}, ErrNotFound
	}
	if food.ID == "" {
		food.ID = newID()
	}
	food.OwnerUserID = userID
	food.Source = "custom"
	m.foodItems[food.ID] = food
	return food, nil
}

func (m *Memory) GetFoodEntry(_ context.Context, userID, entryID string) (FoodEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.foodEntries[userID] {
		if item.ID == entryID {
			return item, nil
		}
	}
	return FoodEntry{}, ErrNotFound
}

func (m *Memory) CreateFoodEntry(_ context.Context, in FoodEntry) (FoodEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[in.UserID]; !ok {
		return FoodEntry{}, ErrNotFound
	}
	if _, ok := m.foodItems[in.FoodID]; !ok {
		return FoodEntry{}, ErrNotFound
	}
	if in.ID == "" {
		in.ID = newID()
	}
	if in.LoggedAt.IsZero() {
		in.LoggedAt = time.Now().UTC()
	}
	m.foodEntries[in.UserID] = append(m.foodEntries[in.UserID], in)
	return in, nil
}

func (m *Memory) ListFoodEntries(_ context.Context, userID string, from, to time.Time) ([]FoodEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]FoodEntry, 0)
	for _, item := range m.foodEntries[userID] {
		if !item.LoggedAt.Before(from) && !item.LoggedAt.After(to) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LoggedAt.Before(out[j].LoggedAt) })
	return out, nil
}

func (m *Memory) DeleteFoodEntry(_ context.Context, userID, entryID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entries := m.foodEntries[userID]
	for i, item := range entries {
		if item.ID == entryID {
			m.foodEntries[userID] = append(entries[:i], entries[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func seedFoodItems() map[string]FoodItem {
	items := []FoodItem{
		{ID: "chicken-breast", Name: "Куриная грудка", Kcal100: 165, Protein100: 31, Fat100: 3.6, Carbs100: 0, Fiber100: 0, ServingG: 150, Source: "seed"},
		{ID: "turkey-breast", Name: "Филе индейки", Kcal100: 135, Protein100: 29, Fat100: 1.6, Carbs100: 0, Fiber100: 0, ServingG: 150, Source: "seed"},
		{ID: "salmon", Name: "Лосось", Kcal100: 208, Protein100: 20, Fat100: 13, Carbs100: 0, Fiber100: 0, ServingG: 150, Source: "seed"},
		{ID: "egg", Name: "Яйцо куриное", Kcal100: 143, Protein100: 12.6, Fat100: 9.5, Carbs100: 0.7, Fiber100: 0, ServingG: 55, Source: "seed"},
		{ID: "cottage-cheese-5", Name: "Творог 5%", Kcal100: 121, Protein100: 17, Fat100: 5, Carbs100: 1.8, Fiber100: 0, ServingG: 200, Source: "seed"},
		{ID: "greek-yogurt", Name: "Йогурт греческий 2%", Kcal100: 73, Protein100: 9.9, Fat100: 2, Carbs100: 3.9, Fiber100: 0, ServingG: 200, Source: "seed"},
		{ID: "oats-dry", Name: "Овсяные хлопья сухие", Kcal100: 379, Protein100: 13.2, Fat100: 6.5, Carbs100: 67.7, Fiber100: 10.1, ServingG: 80, Source: "seed"},
		{ID: "rice-cooked", Name: "Рис белый варёный", Kcal100: 130, Protein100: 2.7, Fat100: 0.3, Carbs100: 28.2, Fiber100: 0.4, ServingG: 200, Source: "seed"},
		{ID: "buckwheat-cooked", Name: "Гречка варёная", Kcal100: 110, Protein100: 4.2, Fat100: 1.1, Carbs100: 21.3, Fiber100: 2.7, ServingG: 200, Source: "seed"},
		{ID: "pasta-cooked", Name: "Макароны варёные", Kcal100: 158, Protein100: 5.8, Fat100: 0.9, Carbs100: 30.9, Fiber100: 1.8, ServingG: 200, Source: "seed"},
		{ID: "potato-boiled", Name: "Картофель варёный", Kcal100: 87, Protein100: 1.9, Fat100: 0.1, Carbs100: 20.1, Fiber100: 1.8, ServingG: 250, Source: "seed"},
		{ID: "banana", Name: "Банан", Kcal100: 89, Protein100: 1.1, Fat100: 0.3, Carbs100: 22.8, Fiber100: 2.6, ServingG: 120, Source: "seed"},
		{ID: "apple", Name: "Яблоко", Kcal100: 52, Protein100: 0.3, Fat100: 0.2, Carbs100: 13.8, Fiber100: 2.4, ServingG: 180, Source: "seed"},
		{ID: "avocado", Name: "Авокадо", Kcal100: 160, Protein100: 2, Fat100: 14.7, Carbs100: 8.5, Fiber100: 6.7, ServingG: 100, Source: "seed"},
		{ID: "olive-oil", Name: "Оливковое масло", Kcal100: 884, Protein100: 0, Fat100: 100, Carbs100: 0, Fiber100: 0, ServingG: 10, Source: "seed"},
		{ID: "almonds", Name: "Миндаль", Kcal100: 579, Protein100: 21.2, Fat100: 49.9, Carbs100: 21.6, Fiber100: 12.5, ServingG: 30, Source: "seed"},
		{ID: "wholegrain-bread", Name: "Хлеб цельнозерновой", Kcal100: 247, Protein100: 13, Fat100: 4.2, Carbs100: 41, Fiber100: 7, ServingG: 60, Source: "seed"},
		{ID: "milk-2", Name: "Молоко 2%", Kcal100: 50, Protein100: 3.3, Fat100: 2, Carbs100: 4.8, Fiber100: 0, ServingG: 250, Source: "seed"},
		{ID: "whey-protein", Name: "Сывороточный протеин", Kcal100: 390, Protein100: 78, Fat100: 6, Carbs100: 8, Fiber100: 0, ServingG: 30, Source: "seed"},
		{ID: "syrniki", Name: "Сырники", Kcal100: 220, Protein100: 14, Fat100: 10, Carbs100: 18, Fiber100: 0.8, ServingG: 200, Source: "seed"},
	}
	out := make(map[string]FoodItem, len(items))
	for _, item := range items {
		out[item.ID] = item
	}
	return out
}

func clonePersonalRecord(in PersonalRecord) PersonalRecord {
	out := in
	out.PreviousValue = cloneFloat(in.PreviousValue)
	return out
}

func valueFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func (m *Memory) detailsLocked(workoutID string) WorkoutDetails {
	return WorkoutDetails{
		Workout:   m.workouts[workoutID],
		Exercises: cloneWorkoutExercises(m.workoutEx[workoutID]),
		Sets:      cloneWorkoutSets(m.workoutSets[workoutID]),
	}
}

func cloneWorkoutExercises(in []WorkoutExercise) []WorkoutExercise {
	out := make([]WorkoutExercise, len(in))
	copy(out, in)
	for i := range out {
		out[i].TargetWeight = cloneFloat(out[i].TargetWeight)
	}
	return out
}

func cloneWorkoutSets(in []WorkoutSet) []WorkoutSet {
	out := make([]WorkoutSet, len(in))
	copy(out, in)
	for i := range out {
		out[i].Weight = cloneFloat(out[i].Weight)
		out[i].RPE = cloneFloat(out[i].RPE)
		out[i].RIR = cloneFloat(out[i].RIR)
		if out[i].CompletedAt != nil {
			t := *out[i].CompletedAt
			out[i].CompletedAt = &t
		}
	}
	return out
}

func cloneFloat(v *float64) *float64 {
	if v == nil {
		return nil
	}
	copy := *v
	return &copy
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func (m *Memory) CreateRecipe(_ context.Context, recipe Recipe) (Recipe, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[recipe.UserID]; !ok {
		return Recipe{}, ErrNotFound
	}
	now := time.Now().UTC()
	if recipe.ID == "" {
		recipe.ID = newID()
	}
	recipe.CreatedAt = now
	recipe.UpdatedAt = now
	recipe.Items = append([]RecipeItem(nil), recipe.Items...)
	m.recipes[recipe.ID] = recipe
	return recipe, nil
}

func (m *Memory) ListRecipes(_ context.Context, userID string) ([]Recipe, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Recipe, 0)
	for _, recipe := range m.recipes {
		if recipe.UserID == userID {
			copy := recipe
			copy.Items = append([]RecipeItem(nil), recipe.Items...)
			out = append(out, copy)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (m *Memory) GetRecipe(_ context.Context, userID, recipeID string) (Recipe, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	recipe, ok := m.recipes[recipeID]
	if !ok {
		return Recipe{}, ErrNotFound
	}
	if recipe.UserID != userID {
		return Recipe{}, ErrForbidden
	}
	recipe.Items = append([]RecipeItem(nil), recipe.Items...)
	return recipe, nil
}

func (m *Memory) CreateBodyMeasurement(_ context.Context, in BodyMeasurement) (BodyMeasurement, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[in.UserID]; !ok {
		return BodyMeasurement{}, ErrNotFound
	}
	if in.ID == "" {
		in.ID = newID()
	}
	if in.LoggedAt.IsZero() {
		in.LoggedAt = time.Now().UTC()
	}
	m.measurements[in.UserID] = append(m.measurements[in.UserID], in)
	return in, nil
}

func (m *Memory) ListBodyMeasurements(_ context.Context, userID string, from, to time.Time) ([]BodyMeasurement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]BodyMeasurement, 0)
	for _, item := range m.measurements[userID] {
		if !item.LoggedAt.Before(from) && !item.LoggedAt.After(to) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LoggedAt.Before(out[j].LoggedAt) })
	return out, nil
}

func (m *Memory) CreateProgram(_ context.Context, program Program, sessions []ProgramSession) (ProgramWithSessions, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[program.UserID]; !ok {
		return ProgramWithSessions{}, ErrNotFound
	}
	for id, existing := range m.programs {
		if existing.UserID == program.UserID && existing.Status == "active" {
			existing.Status = "archived"
			existing.UpdatedAt = time.Now().UTC()
			m.programs[id] = existing
		}
	}
	if program.ID == "" {
		program.ID = newID()
	}
	now := time.Now().UTC()
	if program.CreatedAt.IsZero() {
		program.CreatedAt = now
	}
	program.UpdatedAt = now
	if program.Status == "" {
		program.Status = "active"
	}
	for i := range sessions {
		if sessions[i].ID == "" {
			sessions[i].ID = newID()
		}
		sessions[i].ProgramID = program.ID
	}
	m.programs[program.ID] = program
	m.programSessions[program.ID] = cloneProgramSessions(sessions)
	return ProgramWithSessions{Program: program, Sessions: cloneProgramSessions(sessions)}, nil
}

func (m *Memory) GetProgram(_ context.Context, userID, programID string) (ProgramWithSessions, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.programs[programID]
	if !ok {
		return ProgramWithSessions{}, ErrNotFound
	}
	if p.UserID != userID {
		return ProgramWithSessions{}, ErrForbidden
	}
	return ProgramWithSessions{Program: p, Sessions: cloneProgramSessions(m.programSessions[programID])}, nil
}

func (m *Memory) GetActiveProgram(_ context.Context, userID string) (ProgramWithSessions, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var best *Program
	for _, p := range m.programs {
		if p.UserID != userID || p.Status != "active" {
			continue
		}
		if best == nil || p.UpdatedAt.After(best.UpdatedAt) {
			cp := p
			best = &cp
		}
	}
	if best == nil {
		return ProgramWithSessions{}, ErrNotFound
	}
	return ProgramWithSessions{Program: *best, Sessions: cloneProgramSessions(m.programSessions[best.ID])}, nil
}

func (m *Memory) ListPrograms(_ context.Context, userID string, limit int) ([]ProgramWithSessions, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	items := make([]Program, 0)
	for _, p := range m.programs {
		if p.UserID == userID {
			items = append(items, p)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]ProgramWithSessions, 0, len(items))
	for _, p := range items {
		out = append(out, ProgramWithSessions{Program: p, Sessions: cloneProgramSessions(m.programSessions[p.ID])})
	}
	return out, nil
}

func (m *Memory) GetProgramSession(_ context.Context, userID, sessionID string) (ProgramSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for programID, sessions := range m.programSessions {
		p := m.programs[programID]
		if p.UserID != userID {
			continue
		}
		for _, session := range sessions {
			if session.ID == sessionID {
				return cloneProgramSession(session), nil
			}
		}
	}
	return ProgramSession{}, ErrNotFound
}

func (m *Memory) UpdateProgramSession(_ context.Context, userID string, session ProgramSession) (ProgramSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.programs[session.ProgramID]
	if !ok {
		return ProgramSession{}, ErrNotFound
	}
	if p.UserID != userID {
		return ProgramSession{}, ErrForbidden
	}
	sessions := m.programSessions[session.ProgramID]
	for i := range sessions {
		if sessions[i].ID == session.ID {
			sessions[i] = cloneProgramSession(session)
			m.programSessions[session.ProgramID] = sessions
			p.UpdatedAt = time.Now().UTC()
			m.programs[p.ID] = p
			return cloneProgramSession(session), nil
		}
	}
	return ProgramSession{}, ErrNotFound
}

func (m *Memory) FindProgramSessionByWorkout(_ context.Context, userID, workoutID string) (ProgramSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for programID, sessions := range m.programSessions {
		if m.programs[programID].UserID != userID {
			continue
		}
		for _, session := range sessions {
			if session.WorkoutID != nil && *session.WorkoutID == workoutID {
				return cloneProgramSession(session), nil
			}
		}
	}
	return ProgramSession{}, ErrNotFound
}

func (m *Memory) UpdateProgramStatus(_ context.Context, userID, programID, status string) (ProgramWithSessions, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.programs[programID]
	if !ok {
		return ProgramWithSessions{}, ErrNotFound
	}
	if p.UserID != userID {
		return ProgramWithSessions{}, ErrForbidden
	}
	p.Status = status
	p.UpdatedAt = time.Now().UTC()
	m.programs[programID] = p
	return ProgramWithSessions{Program: p, Sessions: cloneProgramSessions(m.programSessions[programID])}, nil
}

func cloneProgramSession(in ProgramSession) ProgramSession {
	out := in
	if in.WorkoutID != nil {
		v := *in.WorkoutID
		out.WorkoutID = &v
	}
	if in.CompletedAt != nil {
		v := *in.CompletedAt
		out.CompletedAt = &v
	}
	return out
}
func cloneProgramSessions(in []ProgramSession) []ProgramSession {
	out := make([]ProgramSession, len(in))
	for i := range in {
		out[i] = cloneProgramSession(in[i])
	}
	return out
}

func (m *Memory) CreateBodyScan(_ context.Context, scan BodyScan) (BodyScanDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[scan.UserID]; !ok {
		return BodyScanDetails{}, ErrNotFound
	}
	if scan.ID == "" {
		scan.ID = newID()
	}
	if scan.Status == "" {
		scan.Status = "draft"
	}
	if scan.CreatedAt.IsZero() {
		scan.CreatedAt = time.Now().UTC()
	}
	m.bodyScans[scan.ID] = scan
	if _, ok := m.bodyScanPhotos[scan.ID]; !ok {
		m.bodyScanPhotos[scan.ID] = map[string]BodyScanPhoto{}
	}
	return BodyScanDetails{Scan: scan, Photos: []BodyScanPhoto{}}, nil
}

func (m *Memory) GetBodyScan(_ context.Context, userID, scanID string) (BodyScanDetails, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	scan, ok := m.bodyScans[scanID]
	if !ok || scan.UserID != userID {
		return BodyScanDetails{}, ErrNotFound
	}
	photos := make([]BodyScanPhoto, 0, 3)
	for _, view := range []string{"front", "side", "back"} {
		if p, ok := m.bodyScanPhotos[scanID][view]; ok {
			p.QualityIssues = append([]string(nil), p.QualityIssues...)
			photos = append(photos, p)
		}
	}
	return BodyScanDetails{Scan: scan, Photos: photos}, nil
}

func (m *Memory) ListBodyScans(ctx context.Context, userID string, limit int) ([]BodyScanDetails, error) {
	m.mu.RLock()
	ids := make([]string, 0)
	for id, scan := range m.bodyScans {
		if scan.UserID == userID {
			ids = append(ids, id)
		}
	}
	m.mu.RUnlock()
	sort.Slice(ids, func(i, j int) bool {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return m.bodyScans[ids[i]].CreatedAt.After(m.bodyScans[ids[j]].CreatedAt)
	})
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if len(ids) > limit {
		ids = ids[:limit]
	}
	out := make([]BodyScanDetails, 0, len(ids))
	for _, id := range ids {
		d, err := m.GetBodyScan(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (m *Memory) UpsertBodyScanPhoto(ctx context.Context, photo BodyScanPhoto) (BodyScanDetails, error) {
	m.mu.Lock()
	scan, ok := m.bodyScans[photo.ScanID]
	if !ok || scan.UserID != photo.UserID {
		m.mu.Unlock()
		return BodyScanDetails{}, ErrNotFound
	}
	if scan.Status == "completed" {
		m.mu.Unlock()
		return BodyScanDetails{}, ErrInvalidState
	}
	if photo.ID == "" {
		photo.ID = newID()
	}
	if photo.CreatedAt.IsZero() {
		photo.CreatedAt = time.Now().UTC()
	}
	if _, ok := m.bodyScanPhotos[photo.ScanID]; !ok {
		m.bodyScanPhotos[photo.ScanID] = map[string]BodyScanPhoto{}
	}
	m.bodyScanPhotos[photo.ScanID][photo.View] = photo
	m.mu.Unlock()
	return m.GetBodyScan(ctx, photo.UserID, photo.ScanID)
}

func (m *Memory) CompleteBodyScan(ctx context.Context, userID, scanID string, completedAt time.Time) (BodyScanDetails, error) {
	m.mu.Lock()
	scan, ok := m.bodyScans[scanID]
	if !ok || scan.UserID != userID {
		m.mu.Unlock()
		return BodyScanDetails{}, ErrNotFound
	}
	photos := m.bodyScanPhotos[scanID]
	for _, view := range []string{"front", "side", "back"} {
		if _, ok := photos[view]; !ok {
			m.mu.Unlock()
			return BodyScanDetails{}, ErrInvalidState
		}
	}
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	scan.Status = "completed"
	scan.CompletedAt = &completedAt
	m.bodyScans[scanID] = scan
	m.mu.Unlock()
	return m.GetBodyScan(ctx, userID, scanID)
}

func (m *Memory) DeleteBodyScan(_ context.Context, userID, scanID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	scan, ok := m.bodyScans[scanID]
	if !ok || scan.UserID != userID {
		return ErrNotFound
	}
	delete(m.bodyScans, scanID)
	delete(m.bodyScanPhotos, scanID)
	return nil
}

func (m *Memory) SaveTechniqueAnalysis(_ context.Context, a TechniqueAnalysis) (TechniqueAnalysis, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[a.UserID]; !ok {
		return TechniqueAnalysis{}, ErrNotFound
	}
	if a.ID == "" {
		a.ID = newID()
	}
	if a.CaptureMode == "" {
		a.CaptureMode = "recorded"
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	m.techniqueAnalyses[a.ID] = a
	return a, nil
}
func (m *Memory) UpdateTechniqueAnalysisResult(_ context.Context, userID, id, resultJSON string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.techniqueAnalyses[id]
	if !ok || a.UserID != userID {
		return ErrNotFound
	}
	a.ResultJSON = resultJSON
	m.techniqueAnalyses[id] = a
	return nil
}
func (m *Memory) GetTechniqueAnalysis(_ context.Context, userID, id string) (TechniqueAnalysis, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.techniqueAnalyses[id]
	if !ok || a.UserID != userID {
		return TechniqueAnalysis{}, ErrNotFound
	}
	return a, nil
}
func (m *Memory) ListTechniqueAnalyses(_ context.Context, userID string, limit int) ([]TechniqueAnalysis, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	out := []TechniqueAnalysis{}
	for _, a := range m.techniqueAnalyses {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func recoveryKey(userID, localDate string) string { return userID + "|" + localDate }

func cloneSoreness(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (m *Memory) UpsertRecoveryCheckIn(_ context.Context, in RecoveryCheckIn) (RecoveryCheckIn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[in.UserID]; !ok {
		return RecoveryCheckIn{}, ErrNotFound
	}
	key := recoveryKey(in.UserID, in.LocalDate)
	current, exists := m.recoveryCheckIns[key]
	if !exists {
		current.ID = newID()
		current.UserID = in.UserID
		current.LocalDate = in.LocalDate
		current.CreatedAt = time.Now().UTC()
	}
	current.SleepHours = in.SleepHours
	current.SleepQuality = in.SleepQuality
	current.Energy = in.Energy
	current.Stress = in.Stress
	current.MuscleSoreness = cloneSoreness(in.MuscleSoreness)
	current.Notes = in.Notes
	current.UpdatedAt = time.Now().UTC()
	m.recoveryCheckIns[key] = current
	current.MuscleSoreness = cloneSoreness(current.MuscleSoreness)
	return current, nil
}

func (m *Memory) GetRecoveryCheckIn(_ context.Context, userID, localDate string) (RecoveryCheckIn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.recoveryCheckIns[recoveryKey(userID, localDate)]
	if !ok {
		return RecoveryCheckIn{}, ErrNotFound
	}
	item.MuscleSoreness = cloneSoreness(item.MuscleSoreness)
	return item, nil
}

func (m *Memory) ListRecoveryCheckIns(_ context.Context, userID string, limit int) ([]RecoveryCheckIn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 90 {
		limit = 30
	}
	out := make([]RecoveryCheckIn, 0)
	for _, item := range m.recoveryCheckIns {
		if item.UserID == userID {
			item.MuscleSoreness = cloneSoreness(item.MuscleSoreness)
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LocalDate > out[j].LocalDate })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func healthSnapshotKey(userID, localDate, sourcePackage string) string {
	return userID + "|" + localDate + "|" + sourcePackage
}

func healthSourceRank(sourcePackage string) int {
	if sourcePackage == "com.xiaomi.wearable" {
		return 0
	}
	return 10
}

func preferHealthSnapshot(a, b HealthDailySnapshot) HealthDailySnapshot {
	ra, rb := healthSourceRank(a.SourcePackage), healthSourceRank(b.SourcePackage)
	if ra != rb {
		if ra < rb {
			return a
		}
		return b
	}
	if a.ImportedAt.After(b.ImportedAt) {
		return a
	}
	return b
}

func cloneStrings(in []string) []string { return append([]string(nil), in...) }

func (m *Memory) UpsertHealthDailySnapshot(_ context.Context, in HealthDailySnapshot) (HealthDailySnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[in.UserID]; !ok {
		return HealthDailySnapshot{}, ErrNotFound
	}
	key := healthSnapshotKey(in.UserID, in.LocalDate, in.SourcePackage)
	current, exists := m.healthSnapshots[key]
	if !exists {
		current.ID = newID()
		current.UserID = in.UserID
		current.LocalDate = in.LocalDate
	}
	current.Provider = in.Provider
	current.SourcePackage = in.SourcePackage
	current.SourceLabel = in.SourceLabel
	current.Steps = in.Steps
	current.DistanceM = in.DistanceM
	current.ActiveCaloriesKcal = in.ActiveCaloriesKcal
	current.SleepMinutes = in.SleepMinutes
	current.DeepSleepMinutes = in.DeepSleepMinutes
	current.LightSleepMinutes = in.LightSleepMinutes
	current.REMSleepMinutes = in.REMSleepMinutes
	current.AwakeMinutes = in.AwakeMinutes
	current.ExerciseMinutes = in.ExerciseMinutes
	current.ExerciseSessions = in.ExerciseSessions
	current.ExerciseHeartRateAvg = in.ExerciseHeartRateAvg
	current.ExerciseHeartRateMax = in.ExerciseHeartRateMax
	current.RestingHeartRate = in.RestingHeartRate
	current.DataTypes = cloneStrings(in.DataTypes)
	current.CapturedAt = in.CapturedAt
	current.ImportedAt = time.Now().UTC()
	m.healthSnapshots[key] = current
	current.DataTypes = cloneStrings(current.DataTypes)
	return current, nil
}

func (m *Memory) GetHealthDailySnapshot(_ context.Context, userID, localDate string) (HealthDailySnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var selected HealthDailySnapshot
	found := false
	for _, item := range m.healthSnapshots {
		if item.UserID != userID || item.LocalDate != localDate {
			continue
		}
		if !found {
			selected = item
			found = true
		} else {
			selected = preferHealthSnapshot(selected, item)
		}
	}
	if !found {
		return HealthDailySnapshot{}, ErrNotFound
	}
	selected.DataTypes = cloneStrings(selected.DataTypes)
	return selected, nil
}

func (m *Memory) ListHealthDailySnapshotsForDate(_ context.Context, userID, localDate string) ([]HealthDailySnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]HealthDailySnapshot, 0)
	for _, item := range m.healthSnapshots {
		if item.UserID == userID && item.LocalDate == localDate {
			item.DataTypes = cloneStrings(item.DataTypes)
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := healthSourceRank(out[i].SourcePackage), healthSourceRank(out[j].SourcePackage)
		if ri != rj {
			return ri < rj
		}
		return out[i].ImportedAt.After(out[j].ImportedAt)
	})
	return out, nil
}

func (m *Memory) ListHealthDailySnapshots(_ context.Context, userID string, limit int) ([]HealthDailySnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 90 {
		limit = 30
	}
	byDate := map[string]HealthDailySnapshot{}
	for _, item := range m.healthSnapshots {
		if item.UserID != userID {
			continue
		}
		current, ok := byDate[item.LocalDate]
		if !ok {
			byDate[item.LocalDate] = item
		} else {
			byDate[item.LocalDate] = preferHealthSnapshot(current, item)
		}
	}
	out := make([]HealthDailySnapshot, 0, len(byDate))
	for _, item := range byDate {
		item.DataTypes = cloneStrings(item.DataTypes)
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LocalDate > out[j].LocalDate })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
