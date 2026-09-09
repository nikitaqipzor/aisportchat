package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/aifitness"
	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/bodyscan"
	"github.com/example/ai-fitness-os/services/api/internal/catalog"
	"github.com/example/ai-fitness-os/services/api/internal/healthdata"
	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/nutrition"
	"github.com/example/ai-fitness-os/services/api/internal/profile"
	"github.com/example/ai-fitness-os/services/api/internal/programs"
	"github.com/example/ai-fitness-os/services/api/internal/progress"
	"github.com/example/ai-fitness-os/services/api/internal/recovery"
	"github.com/example/ai-fitness-os/services/api/internal/store"
	"github.com/example/ai-fitness-os/services/api/internal/technique"
	"github.com/example/ai-fitness-os/services/api/internal/workouts"
)

type contextKey string

const userIDKey contextKey = "user_id"

type Server struct {
	store             store.Store
	auth              *auth.Service
	tokens            *auth.TokenManager
	workoutService    *workouts.Service
	nutritionService  *nutrition.Service
	progressService   *progress.Service
	programService    *programs.Service
	aiService         *aifitness.Service
	bodyScanService   *bodyscan.Service
	techniqueService  *technique.Service
	recoveryService   *recovery.Service
	healthDataService *healthdata.Service
	authRateLimiter   *authRateLimiter
}

func NewServer() http.Handler {
	st := store.NewMemory()
	tm := auth.NewTokenManager("dev-only-change-me", 15*time.Minute, 30*24*time.Hour)
	return NewServerWithAI(st, tm, aifitness.NewLocalProvider())
}

func NewServerWithDependencies(st store.Store, tm *auth.TokenManager) http.Handler {
	return NewServerWithAI(st, tm, aifitness.NewLocalProvider())
}

func NewServerWithAI(st store.Store, tm *auth.TokenManager, aiProvider aifitness.Provider) http.Handler {
	return NewServerWithAIAndMedia(st, tm, aiProvider, media.NewMemoryStore())
}

func NewServerWithAIAndMedia(st store.Store, tm *auth.TokenManager, aiProvider aifitness.Provider, mediaStore media.Store) http.Handler {
	engine := workouts.NewEngine(catalog.Exercises)
	s := &Server{
		store:             st,
		auth:              auth.NewService(st, tm),
		tokens:            tm,
		workoutService:    workouts.NewService(st, engine),
		nutritionService:  nutrition.NewService(st),
		progressService:   progress.NewService(st),
		programService:    programs.NewService(st),
		bodyScanService:   bodyscan.NewService(st, mediaStore),
		techniqueService:  technique.NewService(st),
		recoveryService:   recovery.NewService(st),
		healthDataService: healthdata.NewService(st),
		authRateLimiter:   newAuthRateLimiter(),
	}
	s.aiService = aifitness.NewService(st, s.nutritionService, s.progressService, aiProvider, s.recoveryService)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)

	mux.HandleFunc("POST /api/v1/auth/register", s.register)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)

	mux.HandleFunc("GET /api/v1/muscles", s.listMuscles)
	mux.Handle("GET /api/v1/muscles/{muscle_id}/stats", s.requireAuth(http.HandlerFunc(s.muscleStats)))
	mux.HandleFunc("GET /api/v1/equipment", s.listEquipment)

	mux.Handle("GET /api/v1/profile", s.requireAuth(http.HandlerFunc(s.getProfile)))
	mux.Handle("PATCH /api/v1/profile", s.requireAuth(http.HandlerFunc(s.updateProfile)))
	mux.Handle("PUT /api/v1/profile/goal", s.requireAuth(http.HandlerFunc(s.setGoal)))
	mux.Handle("PUT /api/v1/profile/training-preferences", s.requireAuth(http.HandlerFunc(s.setTrainingPreferences)))
	mux.Handle("GET /api/v1/onboarding/status", s.requireAuth(http.HandlerFunc(s.getOnboardingStatus)))
	mux.Handle("POST /api/v1/onboarding/complete", s.requireAuth(http.HandlerFunc(s.completeOnboarding)))

	mux.Handle("GET /api/v1/recovery/today", s.requireAuth(http.HandlerFunc(s.recoveryToday)))
	mux.Handle("GET /api/v1/recovery/timeline", s.requireAuth(http.HandlerFunc(s.recoveryTimeline)))
	mux.Handle("PUT /api/v1/recovery/check-in", s.requireAuth(http.HandlerFunc(s.recoveryCheckIn)))
	mux.Handle("GET /api/v1/recovery/history", s.requireAuth(http.HandlerFunc(s.recoveryHistory)))
	mux.Handle("PUT /api/v1/health/snapshots", s.requireAuth(http.HandlerFunc(s.importHealthSnapshot)))
	mux.Handle("GET /api/v1/health/today", s.requireAuth(http.HandlerFunc(s.healthToday)))
	mux.Handle("GET /api/v1/health/history", s.requireAuth(http.HandlerFunc(s.healthHistory)))
	mux.Handle("GET /api/v1/health/insights", s.requireAuth(http.HandlerFunc(s.healthInsights)))

	mux.Handle("POST /api/v1/workouts/generate", s.requireAuth(http.HandlerFunc(s.generateWorkout)))
	mux.Handle("GET /api/v1/workouts/history", s.requireAuth(http.HandlerFunc(s.workoutHistory)))
	mux.Handle("GET /api/v1/records", s.requireAuth(http.HandlerFunc(s.personalRecords)))
	mux.Handle("GET /api/v1/workouts/active", s.requireAuth(http.HandlerFunc(s.activeWorkout)))
	mux.Handle("GET /api/v1/workouts/{workout_id}", s.requireAuth(http.HandlerFunc(s.getWorkout)))
	mux.Handle("PATCH /api/v1/workouts/{workout_id}/favorite", s.requireAuth(http.HandlerFunc(s.favoriteWorkout)))
	mux.Handle("POST /api/v1/workouts/{workout_id}/repeat", s.requireAuth(http.HandlerFunc(s.repeatWorkout)))
	mux.Handle("POST /api/v1/workouts/{workout_id}/start", s.requireAuth(http.HandlerFunc(s.startWorkout)))
	mux.Handle("PUT /api/v1/workouts/{workout_id}/sets", s.requireAuth(http.HandlerFunc(s.logWorkoutSet)))
	mux.Handle("POST /api/v1/workouts/{workout_id}/exercises/{workout_exercise_id}/replace", s.requireAuth(http.HandlerFunc(s.replaceWorkoutExercise)))
	mux.Handle("POST /api/v1/workouts/{workout_id}/finish", s.requireAuth(http.HandlerFunc(s.finishWorkout)))
	mux.Handle("POST /api/v1/workouts/{workout_id}/cancel", s.requireAuth(http.HandlerFunc(s.cancelWorkout)))

	mux.Handle("POST /api/v1/programs/generate", s.requireAuth(http.HandlerFunc(s.generateProgram)))
	mux.Handle("GET /api/v1/programs/active", s.requireAuth(http.HandlerFunc(s.activeProgram)))
	mux.Handle("GET /api/v1/programs/history", s.requireAuth(http.HandlerFunc(s.programHistory)))
	mux.Handle("GET /api/v1/programs/{program_id}", s.requireAuth(http.HandlerFunc(s.getProgram)))
	mux.Handle("GET /api/v1/programs/{program_id}/analytics", s.requireAuth(http.HandlerFunc(s.programAnalytics)))
	mux.Handle("POST /api/v1/programs/{program_id}/archive", s.requireAuth(http.HandlerFunc(s.archiveProgram)))
	mux.Handle("POST /api/v1/programs/sessions/{session_id}/workout", s.requireAuth(http.HandlerFunc(s.programSessionWorkout)))
	mux.Handle("POST /api/v1/programs/sessions/{session_id}/reschedule", s.requireAuth(http.HandlerFunc(s.rescheduleProgramSession)))
	mux.Handle("POST /api/v1/programs/sessions/{session_id}/auto-reschedule", s.requireAuth(http.HandlerFunc(s.autoRescheduleProgramSession)))
	mux.Handle("POST /api/v1/programs/sessions/{session_id}/explain", s.requireAuth(http.HandlerFunc(s.explainProgramSession)))
	mux.Handle("GET /api/v1/nutrition/profile", s.requireAuth(http.HandlerFunc(s.getNutritionProfile)))
	mux.Handle("PUT /api/v1/nutrition/profile", s.requireAuth(http.HandlerFunc(s.setNutritionProfile)))
	mux.Handle("GET /api/v1/nutrition/today", s.requireAuth(http.HandlerFunc(s.nutritionToday)))
	mux.Handle("GET /api/v1/nutrition/history", s.requireAuth(http.HandlerFunc(s.nutritionHistory)))
	mux.Handle("GET /api/v1/nutrition/foods", s.requireAuth(http.HandlerFunc(s.searchFoods)))
	mux.Handle("POST /api/v1/nutrition/entries", s.requireAuth(http.HandlerFunc(s.logFoodEntry)))
	mux.Handle("DELETE /api/v1/nutrition/entries/{entry_id}", s.requireAuth(http.HandlerFunc(s.deleteFoodEntry)))
	mux.Handle("POST /api/v1/nutrition/foods/custom", s.requireAuth(http.HandlerFunc(s.createCustomFood)))
	mux.Handle("GET /api/v1/nutrition/foods/barcode/{barcode}", s.requireAuth(http.HandlerFunc(s.foodByBarcode)))
	mux.Handle("POST /api/v1/nutrition/entries/{entry_id}/repeat", s.requireAuth(http.HandlerFunc(s.repeatFoodEntry)))
	mux.Handle("GET /api/v1/nutrition/recipes", s.requireAuth(http.HandlerFunc(s.listRecipes)))
	mux.Handle("POST /api/v1/nutrition/recipes", s.requireAuth(http.HandlerFunc(s.createRecipe)))
	mux.Handle("POST /api/v1/nutrition/recipes/{recipe_id}/log", s.requireAuth(http.HandlerFunc(s.logRecipe)))
	mux.Handle("GET /api/v1/nutrition/correlation", s.requireAuth(http.HandlerFunc(s.nutritionCorrelation)))
	mux.Handle("POST /api/v1/progress/measurements", s.requireAuth(http.HandlerFunc(s.logMeasurement)))
	mux.Handle("GET /api/v1/progress/measurements", s.requireAuth(http.HandlerFunc(s.measurementHistory)))
	mux.Handle("GET /api/v1/progress/summary", s.requireAuth(http.HandlerFunc(s.progressSummary)))
	mux.Handle("POST /api/v1/body-scans", s.requireAuth(http.HandlerFunc(s.createBodyScan)))
	mux.Handle("GET /api/v1/body-scans", s.requireAuth(http.HandlerFunc(s.listBodyScans)))
	mux.Handle("GET /api/v1/body-scans/comparison/latest", s.requireAuth(http.HandlerFunc(s.latestBodyScanComparison)))
	mux.Handle("GET /api/v1/body-scans/{scan_id}", s.requireAuth(http.HandlerFunc(s.getBodyScan)))
	mux.Handle("DELETE /api/v1/body-scans/{scan_id}", s.requireAuth(http.HandlerFunc(s.deleteBodyScan)))
	mux.Handle("PUT /api/v1/body-scans/{scan_id}/photos/{view}", s.requireAuth(http.HandlerFunc(s.putBodyScanPhoto)))
	mux.Handle("GET /api/v1/body-scans/{scan_id}/photos/{view}", s.requireAuth(http.HandlerFunc(s.getBodyScanPhoto)))
	mux.Handle("POST /api/v1/body-scans/{scan_id}/complete", s.requireAuth(http.HandlerFunc(s.completeBodyScan)))
	mux.Handle("GET /api/v1/technique/exercises", s.requireAuth(http.HandlerFunc(s.techniqueExercises)))
	mux.Handle("POST /api/v1/technique/analyses", s.requireAuth(http.HandlerFunc(s.analyzeTechnique)))
	mux.Handle("GET /api/v1/technique/analyses", s.requireAuth(http.HandlerFunc(s.listTechniqueAnalyses)))
	mux.Handle("GET /api/v1/technique/analyses/{analysis_id}", s.requireAuth(http.HandlerFunc(s.getTechniqueAnalysis)))
	mux.Handle("GET /api/v1/ai/status", s.requireAuth(http.HandlerFunc(s.aiStatus)))
	mux.Handle("POST /api/v1/ai/food/parse", s.requireAuth(http.HandlerFunc(s.aiParseFoodText)))
	mux.Handle("POST /api/v1/ai/food/photo", s.requireAuth(http.HandlerFunc(s.aiParseFoodPhoto)))
	mux.Handle("POST /api/v1/ai/food/confirm", s.requireAuth(http.HandlerFunc(s.aiConfirmFood)))
	mux.Handle("POST /api/v1/ai/chat", s.requireAuth(http.HandlerFunc(s.aiChat)))
	mux.Handle("GET /api/v1/ai/reports/weekly", s.requireAuth(http.HandlerFunc(s.aiWeeklyReport)))
	return requestLogger(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !s.enforceAuthRateLimit(w, r, "register", normalizedIdentity(in.Email), 60, 8, 10*time.Minute) {
		return
	}
	user, tokens, err := s.auth.Register(r.Context(), in.Email, in.Password)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrEmailExists) {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "tokens": tokens})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !s.enforceAuthRateLimit(w, r, "login", normalizedIdentity(in.Email), 30, 8, 5*time.Minute) {
		return
	}
	user, tokens, err := s.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "tokens": tokens})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !s.enforceAuthRateLimit(w, r, "refresh", tokenFingerprint(in.RefreshToken), 90, 15, 5*time.Minute) {
		return
	}
	tokens, err := s.auth.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": tokens})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := s.auth.Logout(r.Context(), in.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listMuscles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": catalog.Muscles})
}

func (s *Server) muscleStats(w http.ResponseWriter, r *http.Request) {
	out, err := s.workoutService.MuscleStats(r.Context(), currentUserID(r.Context()), r.PathValue("muscle_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listEquipment(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": catalog.Equipment})
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	p, _ := s.store.GetProfile(r.Context(), userID)
	goal, goalErr := s.store.GetGoal(r.Context(), userID)
	prefs, prefsErr := s.store.GetTrainingPreferences(r.Context(), userID)
	payload := map[string]any{"profile": p}
	if goalErr == nil {
		payload["goal"] = goal
	}
	if prefsErr == nil {
		payload["training_preferences"] = prefs
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	var in store.Profile
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.HeightCM != nil && (*in.HeightCM < 100 || *in.HeightCM > 250) {
		writeError(w, 422, "height_cm must be between 100 and 250")
		return
	}
	if in.WeightKG != nil && (*in.WeightKG < 30 || *in.WeightKG > 400) {
		writeError(w, 422, "weight_kg must be between 30 and 400")
		return
	}
	if in.ExperienceLevel != nil {
		if err := profile.ValidateLevel(*in.ExperienceLevel); err != nil {
			writeError(w, 422, err.Error())
			return
		}
	}
	if in.UnitSystem != "" && in.UnitSystem != "metric" && in.UnitSystem != "imperial" {
		writeError(w, 422, "unit_system must be metric or imperial")
		return
	}
	in.UserID = currentUserID(r.Context())
	out, err := s.store.UpsertProfile(r.Context(), in)
	if err != nil {
		writeError(w, 500, "profile update failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) setGoal(w http.ResponseWriter, r *http.Request) {
	var in store.Goal
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := profile.ValidateGoal(in.GoalType); err != nil {
		writeError(w, 422, err.Error())
		return
	}
	if in.TargetWeight != nil && (*in.TargetWeight < 30 || *in.TargetWeight > 400) {
		writeError(w, 422, "target_weight_kg must be between 30 and 400")
		return
	}
	in.UserID = currentUserID(r.Context())
	out, err := s.store.SetGoal(r.Context(), in)
	if err != nil {
		writeError(w, 500, "goal update failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) setTrainingPreferences(w http.ResponseWriter, r *http.Request) {
	var in store.TrainingPreferences
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := profile.ValidateEnvironments(in.Environments); err != nil {
		writeError(w, 422, err.Error())
		return
	}
	if in.WorkoutsPerWeek < 1 || in.WorkoutsPerWeek > 14 {
		writeError(w, 422, "workouts_per_week must be between 1 and 14")
		return
	}
	if in.SessionMinutes < 10 || in.SessionMinutes > 180 {
		writeError(w, 422, "session_minutes must be between 10 and 180")
		return
	}
	allowedEq := map[string]bool{}
	for _, e := range catalog.Equipment {
		allowedEq[e.ID] = true
	}
	for _, id := range in.EquipmentIDs {
		if !allowedEq[id] {
			writeError(w, 422, "unsupported equipment_id: "+id)
			return
		}
	}
	in.UserID = currentUserID(r.Context())
	out, err := s.store.SetTrainingPreferences(r.Context(), in)
	if err != nil {
		writeError(w, 500, "training preferences update failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	out, _ := s.store.GetOnboardingStatus(r.Context(), currentUserID(r.Context()))
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) completeOnboarding(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.CompleteOnboarding(r.Context(), currentUserID(r.Context()))
	if errors.Is(err, store.ErrInvalidState) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "onboarding is incomplete", "status": out})
		return
	}
	if err != nil {
		writeError(w, 500, "onboarding completion failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) generateWorkout(w http.ResponseWriter, r *http.Request) {
	var in workouts.GenerateInput
	if !decodeJSON(w, r, &in) {
		return
	}
	userID := currentUserID(r.Context())
	date := strings.TrimSpace(in.LocalDate)
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	vol, intensity, _, adapted := s.recoveryService.AdaptationForMuscle(r.Context(), userID, date, in.Muscle)
	var out workouts.WorkoutView
	var err error
	if adapted {
		out, err = s.workoutService.GenerateAdapted(r.Context(), userID, in, vol, intensity)
	} else {
		out, err = s.workoutService.Generate(r.Context(), userID, in)
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) activeWorkout(w http.ResponseWriter, r *http.Request) {
	out, err := s.workoutService.Active(r.Context(), currentUserID(r.Context()))
	if errors.Is(err, store.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getWorkout(w http.ResponseWriter, r *http.Request) {
	out, err := s.workoutService.Get(r.Context(), currentUserID(r.Context()), r.PathValue("workout_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) favoriteWorkout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Favorite bool `json:"favorite"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.workoutService.Favorite(r.Context(), currentUserID(r.Context()), r.PathValue("workout_id"), in.Favorite)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) repeatWorkout(w http.ResponseWriter, r *http.Request) {
	out, err := s.workoutService.Repeat(r.Context(), currentUserID(r.Context()), r.PathValue("workout_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) startWorkout(w http.ResponseWriter, r *http.Request) {
	out, err := s.workoutService.Start(r.Context(), currentUserID(r.Context()), r.PathValue("workout_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) logWorkoutSet(w http.ResponseWriter, r *http.Request) {
	var in workouts.SetInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.workoutService.LogSet(r.Context(), currentUserID(r.Context()), r.PathValue("workout_id"), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) replaceWorkoutExercise(w http.ResponseWriter, r *http.Request) {
	var in workouts.ReplaceInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.workoutService.Replace(r.Context(), currentUserID(r.Context()), r.PathValue("workout_id"), r.PathValue("workout_exercise_id"), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) finishWorkout(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	workoutID := r.PathValue("workout_id")
	out, err := s.workoutService.Finish(r.Context(), userID, workoutID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	// A standalone workout may have no program session; this association is best-effort.
	_, _ = s.programService.CompleteByWorkout(r.Context(), userID, workoutID)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) cancelWorkout(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	workoutID := r.PathValue("workout_id")
	out, err := s.workoutService.Cancel(r.Context(), userID, workoutID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	// If the workout came from a program session, make that session available again.
	_, _ = s.programService.DetachCancelledWorkout(r.Context(), userID, workoutID)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) workoutHistory(w http.ResponseWriter, r *http.Request) {
	filter := workouts.HistoryFilter{Limit: 20}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			filter.Limit = parsed
		}
	}
	filter.Muscle = strings.TrimSpace(r.URL.Query().Get("muscle"))
	filter.Environment = strings.TrimSpace(r.URL.Query().Get("environment"))
	filter.Status = strings.TrimSpace(r.URL.Query().Get("status"))
	if raw := r.URL.Query().Get("favorite"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			filter.Favorite = &parsed
		}
	}
	out, err := s.workoutService.History(r.Context(), currentUserID(r.Context()), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) personalRecords(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	out, err := s.workoutService.Records(r.Context(), currentUserID(r.Context()), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) generateProgram(w http.ResponseWriter, r *http.Request) {
	var in programs.GenerateInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.programService.Generate(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) activeProgram(w http.ResponseWriter, r *http.Request) {
	out, err := s.programService.Active(r.Context(), currentUserID(r.Context()))
	if errors.Is(err, store.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) programHistory(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	out, err := s.programService.History(r.Context(), currentUserID(r.Context()), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) getProgram(w http.ResponseWriter, r *http.Request) {
	out, err := s.programService.Get(r.Context(), currentUserID(r.Context()), r.PathValue("program_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) programAnalytics(w http.ResponseWriter, r *http.Request) {
	out, err := s.programService.Analytics(r.Context(), currentUserID(r.Context()), r.PathValue("program_id"), time.Now().UTC())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) archiveProgram(w http.ResponseWriter, r *http.Request) {
	out, err := s.programService.Archive(r.Context(), currentUserID(r.Context()), r.PathValue("program_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) programSessionWorkout(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	session, err := s.programService.Session(r.Context(), userID, r.PathValue("session_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if session.WorkoutID != nil {
		out, e := s.workoutService.Get(r.Context(), userID, *session.WorkoutID)
		if e == nil {
			writeJSON(w, http.StatusOK, map[string]any{"session": session, "workout": out})
			return
		}
	}
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	recoveryVolume, recoveryIntensity, _, hasRecovery := s.recoveryService.AdaptationForMuscle(r.Context(), userID, date, session.Muscle)
	volume, intensity := session.VolumeMultiplier, session.IntensityMultiplier
	if hasRecovery {
		volume *= recoveryVolume
		intensity *= recoveryIntensity
	}
	out, err := s.workoutService.GenerateAdapted(r.Context(), userID, workouts.GenerateInput{Muscle: session.Muscle, Environment: session.Environment, LocalDate: date}, volume, intensity)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	updated, err := s.programService.LinkWorkout(r.Context(), userID, session.ID, out.Workout.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"session": updated, "workout": out})
}

func (s *Server) rescheduleProgramSession(w http.ResponseWriter, r *http.Request) {
	var in programs.RescheduleInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.programService.Reschedule(r.Context(), currentUserID(r.Context()), r.PathValue("session_id"), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) autoRescheduleProgramSession(w http.ResponseWriter, r *http.Request) {
	out, err := s.programService.AutoRescheduleMissed(r.Context(), currentUserID(r.Context()), r.PathValue("session_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) explainProgramSession(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	session, err := s.programService.Session(r.Context(), userID, r.PathValue("session_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	message := fmt.Sprintf("Объясни пользователю кратко и без медицинских утверждений, почему его тренировочная программа адаптировала эту сессию. Неделя %d, мышца %s, статус %s, deload=%t, множитель объёма %.2f, интенсивности %.2f, причина из deterministic engine: %s", session.WeekNumber, session.Muscle, session.Status, session.IsDeload, session.VolumeMultiplier, session.IntensityMultiplier, session.AdaptationReason)
	out, err := s.aiService.Chat(r.Context(), userID, message, nil)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session, "explanation": out})
}

func (s *Server) getNutritionProfile(w http.ResponseWriter, r *http.Request) {
	out, err := s.nutritionService.GetProfile(r.Context(), currentUserID(r.Context()))
	if errors.Is(err, store.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) setNutritionProfile(w http.ResponseWriter, r *http.Request) {
	var in nutrition.ProfileInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.nutritionService.SetProfile(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) nutritionToday(w http.ResponseWriter, r *http.Request) {
	at := time.Now().UTC()
	if raw := strings.TrimSpace(r.URL.Query().Get("date")); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
			return
		}
		at = parsed
	}
	out, err := s.nutritionService.Day(r.Context(), currentUserID(r.Context()), at)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) nutritionHistory(w http.ResponseWriter, r *http.Request) {
	days := 7
	if raw := r.URL.Query().Get("days"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			days = n
		}
	}
	out, err := s.nutritionService.History(r.Context(), currentUserID(r.Context()), days, time.Now().UTC())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) searchFoods(w http.ResponseWriter, r *http.Request) {
	limit := 30
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	out, err := s.nutritionService.SearchFoods(r.Context(), currentUserID(r.Context()), r.URL.Query().Get("query"), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) logFoodEntry(w http.ResponseWriter, r *http.Request) {
	var in struct {
		FoodID    string     `json:"food_id"`
		MealType  string     `json:"meal_type"`
		QuantityG float64    `json:"quantity_g"`
		LoggedAt  *time.Time `json:"logged_at,omitempty"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.nutritionService.LogFood(r.Context(), currentUserID(r.Context()), in.FoodID, in.MealType, in.QuantityG, in.LoggedAt)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) deleteFoodEntry(w http.ResponseWriter, r *http.Request) {
	if err := s.nutritionService.DeleteEntry(r.Context(), currentUserID(r.Context()), r.PathValue("entry_id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createCustomFood(w http.ResponseWriter, r *http.Request) {
	var in nutrition.CustomFoodInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.nutritionService.CreateCustomFood(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) foodByBarcode(w http.ResponseWriter, r *http.Request) {
	out, err := s.nutritionService.FoodByBarcode(r.Context(), currentUserID(r.Context()), r.PathValue("barcode"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) repeatFoodEntry(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MealType string     `json:"meal_type,omitempty"`
		LoggedAt *time.Time `json:"logged_at,omitempty"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.nutritionService.RepeatEntry(r.Context(), currentUserID(r.Context()), r.PathValue("entry_id"), in.MealType, in.LoggedAt)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) createRecipe(w http.ResponseWriter, r *http.Request) {
	var in nutrition.RecipeInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.nutritionService.CreateRecipe(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listRecipes(w http.ResponseWriter, r *http.Request) {
	out, err := s.nutritionService.ListRecipes(r.Context(), currentUserID(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) logRecipe(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MealType string     `json:"meal_type"`
		Scale    float64    `json:"scale,omitempty"`
		LoggedAt *time.Time `json:"logged_at,omitempty"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.nutritionService.LogRecipe(r.Context(), currentUserID(r.Context()), r.PathValue("recipe_id"), in.MealType, in.Scale, in.LoggedAt)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) nutritionCorrelation(w http.ResponseWriter, r *http.Request) {
	days := 30
	if raw := r.URL.Query().Get("days"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			days = n
		}
	}
	out, err := s.nutritionService.Correlation(r.Context(), currentUserID(r.Context()), days, time.Now().UTC())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) logMeasurement(w http.ResponseWriter, r *http.Request) {
	var in progress.MeasurementInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.progressService.Log(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) measurementHistory(w http.ResponseWriter, r *http.Request) {
	days := 30
	if raw := r.URL.Query().Get("days"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			days = n
		}
	}
	out, err := s.progressService.History(r.Context(), currentUserID(r.Context()), days, time.Now().UTC())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) progressSummary(w http.ResponseWriter, r *http.Request) {
	days := 30
	if raw := r.URL.Query().Get("days"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			days = n
		}
	}
	out, err := s.progressService.Summary(r.Context(), currentUserID(r.Context()), days, time.Now().UTC())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) aiStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.aiService.ProviderInfo())
}

func (s *Server) aiParseFoodText(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.aiService.ParseFoodText(r.Context(), currentUserID(r.Context()), in.Text)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) aiParseFoodPhoto(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ImageDataURL string `json:"image_data_url"`
	}
	if !decodeJSONLimit(w, r, &in, 8<<20) {
		return
	}
	out, err := s.aiService.ParseFoodImage(r.Context(), currentUserID(r.Context()), in.ImageDataURL)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) aiConfirmFood(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MealType string                      `json:"meal_type"`
		Items    []aifitness.ConfirmFoodItem `json:"items"`
		LoggedAt *time.Time                  `json:"logged_at,omitempty"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.aiService.ConfirmFood(r.Context(), currentUserID(r.Context()), in.MealType, in.Items, in.LoggedAt)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) aiChat(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Message string                  `json:"message"`
		History []aifitness.ChatMessage `json:"history,omitempty"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.aiService.Chat(r.Context(), currentUserID(r.Context()), in.Message, in.History)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) aiWeeklyReport(w http.ResponseWriter, r *http.Request) {
	out, err := s.aiService.Weekly(r.Context(), currentUserID(r.Context()), time.Now().UTC())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "bearer token required")
			return
		}
		claims, err := s.tokens.ParseAccessToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired access token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.Sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func currentUserID(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return decodeJSONLimit(w, r, dst, 1<<20)
}

func decodeJSONLimit(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, store.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, store.ErrInvalidState):
		writeError(w, http.StatusConflict, "workout state does not allow this operation")
	default:
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(started))
	})
}

func (s *Server) createBodyScan(w http.ResponseWriter, r *http.Request) {
	out, err := s.bodyScanService.Create(r.Context(), currentUserID(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listBodyScans(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}
	out, err := s.bodyScanService.List(r.Context(), currentUserID(r.Context()), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) getBodyScan(w http.ResponseWriter, r *http.Request) {
	out, err := s.bodyScanService.Get(r.Context(), currentUserID(r.Context()), r.PathValue("scan_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) putBodyScanPhoto(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ImageDataURL string `json:"image_data_url"`
	}
	if !decodeJSONLimit(w, r, &in, 12<<20) {
		return
	}
	out, err := s.bodyScanService.AddPhoto(r.Context(), currentUserID(r.Context()), r.PathValue("scan_id"), r.PathValue("view"), in.ImageDataURL)
	if err != nil {
		if errors.Is(err, bodyscan.ErrInvalidImage) || errors.Is(err, bodyscan.ErrQuality) || errors.Is(err, store.ErrInvalidState) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getBodyScanPhoto(w http.ResponseWriter, r *http.Request) {
	blob, err := s.bodyScanService.Photo(r.Context(), currentUserID(r.Context()), r.PathValue("scan_id"), r.PathValue("view"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", blob.MimeType)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(blob.Bytes)
}

func (s *Server) completeBodyScan(w http.ResponseWriter, r *http.Request) {
	out, err := s.bodyScanService.Complete(r.Context(), currentUserID(r.Context()), r.PathValue("scan_id"))
	if err != nil {
		if errors.Is(err, store.ErrInvalidState) {
			writeError(w, http.StatusBadRequest, "front, side and back photos are required")
			return
		}
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) latestBodyScanComparison(w http.ResponseWriter, r *http.Request) {
	out, err := s.bodyScanService.LatestComparison(r.Context(), currentUserID(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deleteBodyScan(w http.ResponseWriter, r *http.Request) {
	if err := s.bodyScanService.Delete(r.Context(), currentUserID(r.Context()), r.PathValue("scan_id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) techniqueExercises(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": technique.SupportedExercises(), "algorithm_version": technique.AlgorithmVersion})
}

func (s *Server) analyzeTechnique(w http.ResponseWriter, r *http.Request) {
	var in technique.AnalyzeInput
	if !decodeJSONLimit(w, r, &in, 6<<20) {
		return
	}
	out, err := s.techniqueService.Analyze(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listTechniqueAnalyses(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := s.techniqueService.List(r.Context(), currentUserID(r.Context()), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) getTechniqueAnalysis(w http.ResponseWriter, r *http.Request) {
	out, err := s.techniqueService.Get(r.Context(), currentUserID(r.Context()), r.PathValue("analysis_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func recoveryDate(r *http.Request) (string, error) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", errors.New("date must be YYYY-MM-DD")
	}
	return date, nil
}

func (s *Server) recoveryToday(w http.ResponseWriter, r *http.Request) {
	date, err := recoveryDate(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := s.recoveryService.Summary(r.Context(), currentUserID(r.Context()), date)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) recoveryTimeline(w http.ResponseWriter, r *http.Request) {
	date, err := recoveryDate(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := s.recoveryService.Timeline(r.Context(), currentUserID(r.Context()), date)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) recoveryCheckIn(w http.ResponseWriter, r *http.Request) {
	var in recovery.CheckInInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Date) == "" {
		in.Date = time.Now().UTC().Format("2006-01-02")
	}
	out, err := s.recoveryService.SaveCheckIn(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) recoveryHistory(w http.ResponseWriter, r *http.Request) {
	limit := 30
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, e := strconv.Atoi(raw); e == nil {
			limit = n
		}
	}
	items, err := s.store.ListRecoveryCheckIns(r.Context(), currentUserID(r.Context()), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) importHealthSnapshot(w http.ResponseWriter, r *http.Request) {
	var in healthdata.SnapshotInput
	if !decodeJSON(w, r, &in) {
		return
	}
	out, err := s.healthDataService.Import(r.Context(), currentUserID(r.Context()), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) healthToday(w http.ResponseWriter, r *http.Request) {
	date, err := recoveryDate(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := s.healthDataService.Get(r.Context(), currentUserID(r.Context()), date)
	if errors.Is(err, store.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) healthInsights(w http.ResponseWriter, r *http.Request) {
	date, err := recoveryDate(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := s.healthDataService.Insights(r.Context(), currentUserID(r.Context()), date)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) healthHistory(w http.ResponseWriter, r *http.Request) {
	limit := 14
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, e := strconv.Atoi(raw); e == nil {
			limit = n
		}
	}
	items, err := s.healthDataService.History(r.Context(), currentUserID(r.Context()), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
