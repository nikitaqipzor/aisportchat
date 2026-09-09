package programs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type Service struct{ store store.Store }

func NewService(st store.Store) *Service { return &Service{store: st} }

type GenerateInput struct {
	Weeks           int        `json:"weeks"`
	WorkoutsPerWeek int        `json:"workouts_per_week,omitempty"`
	Environment     string     `json:"environment,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
}

type Analytics struct {
	ProgramID             string                `json:"program_id"`
	PlannedSessions       int                   `json:"planned_sessions"`
	CompletedSessions     int                   `json:"completed_sessions"`
	MissedSessions        int                   `json:"missed_sessions"`
	RescheduledSessions   int                   `json:"rescheduled_sessions"`
	UpcomingSessions      int                   `json:"upcoming_sessions"`
	DeloadSessions        int                   `json:"deload_sessions"`
	AdherencePercent      float64               `json:"adherence_percent"`
	CurrentWeek           int                   `json:"current_week"`
	TotalWeeks            int                   `json:"total_weeks"`
	PlannedSetsByMuscle   map[string]int        `json:"planned_sets_by_muscle"`
	CompletedSetsByMuscle map[string]int        `json:"completed_sets_by_muscle"`
	NextSession           *store.ProgramSession `json:"next_session,omitempty"`
}

type RescheduleInput struct {
	NewDate time.Time `json:"new_date"`
	Reason  string    `json:"reason,omitempty"`
}

func (s *Service) Generate(ctx context.Context, userID string, in GenerateInput) (store.ProgramWithSessions, error) {
	if in.Weeks != 4 && in.Weeks != 8 && in.Weeks != 12 {
		return store.ProgramWithSessions{}, errors.New("weeks must be one of 4, 8, 12")
	}
	goal, err := s.store.GetGoal(ctx, userID)
	if err != nil {
		return store.ProgramWithSessions{}, err
	}
	prefs, err := s.store.GetTrainingPreferences(ctx, userID)
	if err != nil {
		return store.ProgramWithSessions{}, err
	}

	workoutsPerWeek := in.WorkoutsPerWeek
	if workoutsPerWeek == 0 {
		workoutsPerWeek = prefs.WorkoutsPerWeek
	}
	if workoutsPerWeek < 1 {
		workoutsPerWeek = 1
	}
	if workoutsPerWeek > 7 {
		workoutsPerWeek = 7
	}

	environment := strings.TrimSpace(in.Environment)
	if environment == "" && len(prefs.Environments) > 0 {
		environment = prefs.Environments[0]
	}
	if environment != "home" && environment != "gym" && environment != "band" {
		return store.ProgramWithSessions{}, errors.New("environment must be home, gym or band")
	}
	allowed := false
	for _, item := range prefs.Environments {
		if item == environment {
			allowed = true
			break
		}
	}
	if !allowed {
		return store.ProgramWithSessions{}, errors.New("environment is not enabled in training preferences")
	}

	requestedStart := startOfDay(time.Now().UTC())
	if in.StartDate != nil {
		requestedStart = startOfDay(in.StartDate.UTC())
	}
	// A program owns complete Monday-Sunday weeks. Starting mid-week would create a
	// partial first week and can overlap shifted sessions with week 2, so start on
	// the nearest Monday on/after the requested date.
	monday := requestedStart.AddDate(0, 0, -weekdayOffset(requestedStart.Weekday()))
	if monday.Before(requestedStart) {
		monday = monday.AddDate(0, 0, 7)
	}
	start := monday
	split := splitFor(goal.GoalType, workoutsPerWeek)
	weekdays := trainingWeekdays(workoutsPerWeek)

	sessions := make([]store.ProgramSession, 0, in.Weeks*workoutsPerWeek)
	for week := 1; week <= in.Weeks; week++ {
		deload := week%4 == 0
		for dayIndex := 0; dayIndex < workoutsPerWeek; dayIndex++ {
			planned := monday.AddDate(0, 0, (week-1)*7+weekdays[dayIndex])
			focus := split[dayIndex%len(split)]
			sets := normalSetsForMuscle(focus[0])
			volumeMultiplier := 1.0
			intensityMultiplier := 1.0
			reason := ""
			if deload {
				volumeMultiplier = 0.6
				intensityMultiplier = 0.9
				sets = int(math.Max(4, math.Round(float64(sets)*volumeMultiplier)))
				reason = "Плановая разгрузочная неделя: объём снижен примерно на 40%, интенсивность — примерно на 10%."
			}
			session := store.ProgramSession{
				WeekNumber: week, DayIndex: dayIndex + 1, PlannedDate: planned, OriginalDate: planned,
				Muscle: focus[0], Environment: environment, Status: "planned", IsDeload: deload,
				VolumeMultiplier: volumeMultiplier, IntensityMultiplier: intensityMultiplier,
				PlannedSets: sets, AdaptationReason: reason,
			}
			if len(focus) > 1 {
				session.SecondaryMuscle = focus[1]
			}
			sessions = append(sessions, session)
		}
	}
	title := fmt.Sprintf("%d недель · %d тренировок/нед", in.Weeks, workoutsPerWeek)
	program := store.Program{UserID: userID, Title: title, GoalType: goal.GoalType, Weeks: in.Weeks, WorkoutsPerWeek: workoutsPerWeek, Environment: environment, Status: "active", StartDate: start}
	return s.store.CreateProgram(ctx, program, sessions)
}

func (s *Service) Active(ctx context.Context, userID string) (store.ProgramWithSessions, error) {
	p, err := s.store.GetActiveProgram(ctx, userID)
	if err != nil {
		return p, err
	}
	return s.refreshMissed(ctx, userID, p, time.Now().UTC())
}
func (s *Service) Get(ctx context.Context, userID, programID string) (store.ProgramWithSessions, error) {
	p, err := s.store.GetProgram(ctx, userID, programID)
	if err != nil {
		return p, err
	}
	return s.refreshMissed(ctx, userID, p, time.Now().UTC())
}
func (s *Service) History(ctx context.Context, userID string, limit int) ([]store.ProgramWithSessions, error) {
	return s.store.ListPrograms(ctx, userID, limit)
}

func (s *Service) Session(ctx context.Context, userID, sessionID string) (store.ProgramSession, error) {
	return s.store.GetProgramSession(ctx, userID, sessionID)
}

func (s *Service) LinkWorkout(ctx context.Context, userID, sessionID, workoutID string) (store.ProgramSession, error) {
	session, err := s.store.GetProgramSession(ctx, userID, sessionID)
	if err != nil {
		return session, err
	}
	if session.Status == "completed" {
		return session, store.ErrInvalidState
	}
	session.WorkoutID = &workoutID
	if session.Status == "missed" {
		session.Status = "rescheduled"
	}
	if session.AdaptationReason == "" {
		session.AdaptationReason = "Тренировка создана по календарю программы."
	}
	return s.store.UpdateProgramSession(ctx, userID, session)
}

func (s *Service) CompleteByWorkout(ctx context.Context, userID, workoutID string) (store.ProgramSession, error) {
	session, err := s.store.FindProgramSessionByWorkout(ctx, userID, workoutID)
	if err != nil {
		return session, err
	}
	now := time.Now().UTC()
	session.Status = "completed"
	session.CompletedAt = &now
	if session.AdaptationReason == "" {
		session.AdaptationReason = "Выполнено по плану."
	}
	updated, err := s.store.UpdateProgramSession(ctx, userID, session)
	if err != nil {
		return updated, err
	}
	if program, e := s.store.GetProgram(ctx, userID, session.ProgramID); e == nil {
		allDone := true
		for _, item := range program.Sessions {
			if item.Status != "completed" && item.Status != "skipped" {
				allDone = false
				break
			}
		}
		if allDone {
			_, _ = s.store.UpdateProgramStatus(ctx, userID, session.ProgramID, "completed")
		}
	}
	return updated, nil
}

func (s *Service) DetachCancelledWorkout(ctx context.Context, userID, workoutID string) (store.ProgramSession, error) {
	session, err := s.store.FindProgramSessionByWorkout(ctx, userID, workoutID)
	if err != nil {
		return session, err
	}
	if session.Status == "completed" {
		return session, store.ErrInvalidState
	}
	session.WorkoutID = nil
	session.CompletedAt = nil
	session.AdaptationReason = "Пользователь отменил тренировку. Сессия остаётся в календаре и может быть создана заново."
	return s.store.UpdateProgramSession(ctx, userID, session)
}

func (s *Service) Reschedule(ctx context.Context, userID, sessionID string, in RescheduleInput) (store.ProgramSession, error) {
	session, err := s.store.GetProgramSession(ctx, userID, sessionID)
	if err != nil {
		return session, err
	}
	if session.Status == "completed" {
		return session, store.ErrInvalidState
	}
	if in.NewDate.IsZero() {
		return session, errors.New("new_date is required")
	}
	newDate := startOfDay(in.NewDate.UTC())
	if newDate.Before(startOfDay(time.Now().UTC()).AddDate(0, 0, -1)) {
		return session, errors.New("new_date cannot be in the past")
	}
	session.PlannedDate = newDate
	session.Status = "rescheduled"
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "Пользователь перенёс тренировку."
	}
	session.AdaptationReason = reason
	return s.store.UpdateProgramSession(ctx, userID, session)
}

func (s *Service) AutoRescheduleMissed(ctx context.Context, userID, sessionID string) (store.ProgramSession, error) {
	current, err := s.store.GetProgramSession(ctx, userID, sessionID)
	if err != nil {
		return current, err
	}
	if current.Status != "missed" {
		return current, store.ErrInvalidState
	}
	if current.WorkoutID != nil {
		details, e := s.store.GetWorkout(ctx, userID, *current.WorkoutID)
		if e != nil || details.Workout.Status != "planned" {
			return current, store.ErrInvalidState
		}
	}
	active, err := s.store.GetProgram(ctx, userID, current.ProgramID)
	if err != nil {
		return current, err
	}
	occupied := map[string]bool{}
	for _, x := range active.Sessions {
		if x.ID != current.ID && (x.Status == "planned" || x.Status == "rescheduled") {
			occupied[x.PlannedDate.Format("2006-01-02")] = true
		}
	}
	date := startOfDay(time.Now().UTC()).AddDate(0, 0, 1)
	programEnd := active.Program.StartDate.AddDate(0, 0, active.Program.Weeks*7+6)
	for !date.After(programEnd) {
		if !occupied[date.Format("2006-01-02")] {
			current.PlannedDate = date
			current.Status = "rescheduled"
			current.AdaptationReason = "Автоперенос после пропуска на ближайший свободный день программы."
			return s.store.UpdateProgramSession(ctx, userID, current)
		}
		date = date.AddDate(0, 0, 1)
	}
	return current, errors.New("no free date remains inside the program")
}

func (s *Service) Analytics(ctx context.Context, userID, programID string, now time.Time) (Analytics, error) {
	p, err := s.Get(ctx, userID, programID)
	if err != nil {
		return Analytics{}, err
	}
	out := Analytics{ProgramID: p.Program.ID, TotalWeeks: p.Program.Weeks, PlannedSetsByMuscle: map[string]int{}, CompletedSetsByMuscle: map[string]int{}}
	elapsedDays := int(startOfDay(now.UTC()).Sub(startOfDay(p.Program.StartDate)).Hours() / 24)
	out.CurrentWeek = elapsedDays/7 + 1
	if out.CurrentWeek < 1 {
		out.CurrentWeek = 1
	}
	if out.CurrentWeek > p.Program.Weeks {
		out.CurrentWeek = p.Program.Weeks
	}
	for _, session := range p.Sessions {
		out.PlannedSessions++
		out.PlannedSetsByMuscle[session.Muscle] += session.PlannedSets
		if session.SecondaryMuscle != "" {
			out.PlannedSetsByMuscle[session.SecondaryMuscle] += int(math.Round(float64(session.PlannedSets) * 0.35))
		}
		if session.IsDeload {
			out.DeloadSessions++
		}
		switch session.Status {
		case "completed":
			out.CompletedSessions++
			completedSets := session.PlannedSets
			if session.WorkoutID != nil {
				if details, e := s.store.GetWorkout(ctx, userID, *session.WorkoutID); e == nil {
					completedSets = len(details.Sets)
				}
			}
			out.CompletedSetsByMuscle[session.Muscle] += completedSets
			if session.SecondaryMuscle != "" {
				out.CompletedSetsByMuscle[session.SecondaryMuscle] += int(math.Round(float64(completedSets) * 0.35))
			}
		case "missed":
			out.MissedSessions++
		case "rescheduled":
			out.RescheduledSessions++
			if !session.PlannedDate.Before(startOfDay(now.UTC())) {
				out.UpcomingSessions++
			}
		default:
			if !session.PlannedDate.Before(startOfDay(now.UTC())) {
				out.UpcomingSessions++
			}
		}
		if session.Status != "completed" && !session.PlannedDate.Before(startOfDay(now.UTC())) {
			if out.NextSession == nil || session.PlannedDate.Before(out.NextSession.PlannedDate) {
				cp := session
				out.NextSession = &cp
			}
		}
	}
	denominator := out.CompletedSessions + out.MissedSessions
	if denominator > 0 {
		out.AdherencePercent = math.Round((float64(out.CompletedSessions)/float64(denominator))*1000) / 10
	}
	return out, nil
}

func (s *Service) Archive(ctx context.Context, userID, programID string) (store.ProgramWithSessions, error) {
	return s.store.UpdateProgramStatus(ctx, userID, programID, "archived")
}

func (s *Service) refreshMissed(ctx context.Context, userID string, p store.ProgramWithSessions, now time.Time) (store.ProgramWithSessions, error) {
	today := startOfDay(now.UTC())
	changed := false
	for i := range p.Sessions {
		session := p.Sessions[i]
		if (session.Status == "planned" || session.Status == "rescheduled") && session.PlannedDate.Before(today) {
			canMarkMissed := session.WorkoutID == nil
			if session.WorkoutID != nil {
				if details, e := s.store.GetWorkout(ctx, userID, *session.WorkoutID); e == nil {
					switch details.Workout.Status {
					case "completed":
						now := details.Workout.CompletedAt
						session.Status = "completed"
						session.CompletedAt = now
						session.AdaptationReason = "Завершённая тренировка синхронизирована с программой."
					case "planned":
						canMarkMissed = true
					}
				}
			}
			if canMarkMissed {
				session.Status = "missed"
				session.AdaptationReason = "Запланированная дата прошла без завершённой тренировки."
			}
			if canMarkMissed || session.Status == "completed" {
				updated, err := s.store.UpdateProgramSession(ctx, userID, session)
				if err != nil {
					return p, err
				}
				p.Sessions[i] = updated
				changed = true
			}
		}
	}
	if changed {
		sort.Slice(p.Sessions, func(i, j int) bool { return p.Sessions[i].PlannedDate.Before(p.Sessions[j].PlannedDate) })
	}
	return p, nil
}

func splitFor(goal string, n int) [][]string {
	mass := [][]string{{"chest", "triceps"}, {"back", "biceps"}, {"quads", "glutes"}, {"shoulders", "triceps"}, {"hamstrings", "glutes"}, {"biceps", "forearms"}, {"core", "calves"}}
	strength := [][]string{{"chest", "triceps"}, {"quads", "glutes"}, {"back", "biceps"}, {"shoulders", "triceps"}, {"hamstrings", "glutes"}, {"chest", "triceps"}, {"core", "calves"}}
	general := [][]string{{"back", "biceps"}, {"quads", "glutes"}, {"chest", "triceps"}, {"core", "shoulders"}, {"hamstrings", "glutes"}, {"shoulders", "forearms"}, {"calves", "core"}}
	base := mass
	switch goal {
	case "strength", "increase_strength":
		base = strength
	case "fat_loss", "weight_loss", "maintenance", "maintain", "endurance":
		base = general
	}
	if n > len(base) {
		n = len(base)
	}
	return base[:n]
}
func trainingWeekdays(n int) []int {
	switch n {
	case 1:
		return []int{2}
	case 2:
		return []int{0, 3}
	case 3:
		return []int{0, 2, 4}
	case 4:
		return []int{0, 1, 3, 5}
	case 5:
		return []int{0, 1, 2, 4, 5}
	case 6:
		return []int{0, 1, 2, 3, 4, 5}
	default:
		return []int{0, 1, 2, 3, 4, 5, 6}
	}
}
func normalSetsForMuscle(m string) int {
	switch m {
	case "chest", "back", "quads", "hamstrings", "glutes":
		return 12
	case "shoulders":
		return 10
	default:
		return 8
	}
}
func weekdayOffset(w time.Weekday) int {
	if w == time.Sunday {
		return 6
	}
	return int(w) - 1
}
func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
