package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrEmailExists  = errors.New("email already exists")
	ErrInvalidState = errors.New("invalid state")
	ErrForbidden    = errors.New("forbidden")
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Profile struct {
	UserID          string    `json:"user_id"`
	BirthDate       *string   `json:"birth_date,omitempty"`
	Gender          *string   `json:"gender,omitempty"`
	HeightCM        *float64  `json:"height_cm,omitempty"`
	WeightKG        *float64  `json:"weight_kg,omitempty"`
	ExperienceLevel *string   `json:"experience_level,omitempty"`
	AgeYears        *int      `json:"age_years,omitempty"`
	Injuries        []string  `json:"injuries"`
	Limitations     []string  `json:"limitations"`
	UnitSystem      string    `json:"unit_system"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Goal struct {
	UserID       string    `json:"user_id"`
	GoalType     string    `json:"goal_type"`
	TargetWeight *float64  `json:"target_weight_kg,omitempty"`
	StartedAt    time.Time `json:"started_at"`
}

type TrainingPreferences struct {
	UserID          string    `json:"user_id"`
	Environments    []string  `json:"environments"`
	EquipmentIDs    []string  `json:"equipment_ids"`
	WorkoutsPerWeek int       `json:"workouts_per_week"`
	SessionMinutes  int       `json:"session_minutes"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OnboardingStatus struct {
	ProfileCompleted  bool `json:"profile_completed"`
	GoalCompleted     bool `json:"goal_completed"`
	TrainingCompleted bool `json:"training_completed"`
	Completed         bool `json:"completed"`
}

type RefreshSession struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type Workout struct {
	ID                string     `json:"id"`
	UserID            string     `json:"-"`
	Muscle            string     `json:"muscle"`
	Environment       string     `json:"environment"`
	Status            string     `json:"status"`
	DurationMinutes   int        `json:"duration_minutes"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	DurationSeconds   *int       `json:"duration_seconds,omitempty"`
	TotalVolume       float64    `json:"total_volume"`
	CompletionPercent float64    `json:"completion_percent"`
	EndedEarly        bool       `json:"ended_early"`
	Favorite          bool       `json:"favorite"`
	CreatedAt         time.Time  `json:"created_at"`
}

type WorkoutExercise struct {
	ID              string   `json:"id"`
	WorkoutID       string   `json:"workout_id"`
	ExerciseID      string   `json:"exercise_id"`
	Position        int      `json:"position"`
	TargetSets      int      `json:"target_sets"`
	TargetRepsMin   int      `json:"target_reps_min"`
	TargetRepsMax   int      `json:"target_reps_max"`
	TargetWeight    *float64 `json:"target_weight,omitempty"`
	RestSeconds     int      `json:"rest_seconds"`
	ProgressionNote string   `json:"progression_note,omitempty"`
}

type WorkoutSet struct {
	ID                string     `json:"id"`
	WorkoutExerciseID string     `json:"workout_exercise_id"`
	SetNumber         int        `json:"set_number"`
	Weight            *float64   `json:"weight,omitempty"`
	Repetitions       int        `json:"repetitions"`
	RPE               *float64   `json:"rpe,omitempty"`
	RIR               *float64   `json:"rir,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type WorkoutDetails struct {
	Workout   Workout           `json:"workout"`
	Exercises []WorkoutExercise `json:"exercises"`
	Sets      []WorkoutSet      `json:"sets"`
}

type ExercisePerformance struct {
	ExerciseID  string    `json:"exercise_id"`
	Weight      *float64  `json:"weight,omitempty"`
	Repetitions int       `json:"repetitions"`
	RPE         *float64  `json:"rpe,omitempty"`
	RIR         *float64  `json:"rir,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

type PersonalRecord struct {
	ID            string    `json:"id"`
	UserID        string    `json:"-"`
	WorkoutID     string    `json:"workout_id"`
	WorkoutSetID  string    `json:"workout_set_id,omitempty"`
	ExerciseID    string    `json:"exercise_id"`
	RecordType    string    `json:"record_type"`
	Value         float64   `json:"value"`
	PreviousValue *float64  `json:"previous_value,omitempty"`
	AchievedAt    time.Time `json:"achieved_at"`
}

type NutritionProfile struct {
	UserID          string    `json:"user_id"`
	Goal            string    `json:"goal"`
	ActivityLevel   string    `json:"activity_level"`
	CalculationMode string    `json:"calculation_mode"`
	CalorieTarget   int       `json:"calorie_target"`
	ProteinTarget   float64   `json:"protein_target_g"`
	FatTarget       float64   `json:"fat_target_g"`
	CarbTarget      float64   `json:"carb_target_g"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type FoodItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Brand       string  `json:"brand,omitempty"`
	Barcode     string  `json:"barcode,omitempty"`
	OwnerUserID string  `json:"-"`
	Kcal100     float64 `json:"kcal_per_100g"`
	Protein100  float64 `json:"protein_per_100g"`
	Fat100      float64 `json:"fat_per_100g"`
	Carbs100    float64 `json:"carbs_per_100g"`
	Fiber100    float64 `json:"fiber_per_100g"`
	ServingG    float64 `json:"serving_g"`
	Source      string  `json:"source"`
}

type Recipe struct {
	ID        string       `json:"id"`
	UserID    string       `json:"-"`
	Name      string       `json:"name"`
	Items     []RecipeItem `json:"items"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type RecipeItem struct {
	FoodID    string  `json:"food_id"`
	FoodName  string  `json:"food_name"`
	QuantityG float64 `json:"quantity_g"`
	Calories  float64 `json:"calories"`
	Protein   float64 `json:"protein_g"`
	Fat       float64 `json:"fat_g"`
	Carbs     float64 `json:"carbs_g"`
}

type BodyMeasurement struct {
	ID         string    `json:"id"`
	UserID     string    `json:"-"`
	LoggedAt   time.Time `json:"logged_at"`
	WeightKG   *float64  `json:"weight_kg,omitempty"`
	WaistCM    *float64  `json:"waist_cm,omitempty"`
	ChestCM    *float64  `json:"chest_cm,omitempty"`
	ShoulderCM *float64  `json:"shoulder_cm,omitempty"`
	ArmCM      *float64  `json:"arm_cm,omitempty"`
	ForearmCM  *float64  `json:"forearm_cm,omitempty"`
	HipCM      *float64  `json:"hip_cm,omitempty"`
	ThighCM    *float64  `json:"thigh_cm,omitempty"`
	CalfCM     *float64  `json:"calf_cm,omitempty"`
	NeckCM     *float64  `json:"neck_cm,omitempty"`
}

type FoodEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	FoodID    string    `json:"food_id"`
	FoodName  string    `json:"food_name"`
	MealType  string    `json:"meal_type"`
	LoggedAt  time.Time `json:"logged_at"`
	QuantityG float64   `json:"quantity_g"`
	Calories  float64   `json:"calories"`
	Protein   float64   `json:"protein_g"`
	Fat       float64   `json:"fat_g"`
	Carbs     float64   `json:"carbs_g"`
	Fiber     float64   `json:"fiber_g"`
}

type RecoveryCheckIn struct {
	ID             string         `json:"id"`
	UserID         string         `json:"-"`
	LocalDate      string         `json:"date"`
	SleepHours     float64        `json:"sleep_hours"`
	SleepQuality   int            `json:"sleep_quality"`
	Energy         int            `json:"energy"`
	Stress         int            `json:"stress"`
	MuscleSoreness map[string]int `json:"muscle_soreness"`
	Notes          string         `json:"notes,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type BodyScan struct {
	ID          string     `json:"id"`
	UserID      string     `json:"-"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type BodyScanPhoto struct {
	ID            string    `json:"id"`
	ScanID        string    `json:"scan_id"`
	UserID        string    `json:"-"`
	View          string    `json:"view"`
	StorageKey    string    `json:"-"`
	MimeType      string    `json:"mime_type"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	Bytes         int       `json:"bytes"`
	Brightness    float64   `json:"brightness"`
	Contrast      float64   `json:"contrast"`
	QualityStatus string    `json:"quality_status"`
	QualityIssues []string  `json:"quality_issues,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type BodyScanDetails struct {
	Scan   BodyScan        `json:"scan"`
	Photos []BodyScanPhoto `json:"photos"`
}

type TechniqueAnalysis struct {
	ID                string    `json:"id"`
	UserID            string    `json:"-"`
	ExerciseKey       string    `json:"exercise_key"`
	CaptureMode       string    `json:"capture_mode"`
	WorkoutID         string    `json:"workout_id,omitempty"`
	WorkoutExerciseID string    `json:"workout_exercise_id,omitempty"`
	SetNumber         int       `json:"set_number,omitempty"`
	RepCount          int       `json:"rep_count"`
	TechniqueScore    int       `json:"technique_score"`
	ROMScore          int       `json:"rom_score"`
	TempoScore        int       `json:"tempo_score"`
	SymmetryScore     int       `json:"symmetry_score"`
	StabilityScore    int       `json:"stability_score"`
	Confidence        float64   `json:"confidence"`
	DurationMS        int64     `json:"duration_ms"`
	AlgorithmVersion  string    `json:"algorithm_version"`
	ResultJSON        string    `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
}

type Program struct {
	ID              string    `json:"id"`
	UserID          string    `json:"-"`
	Title           string    `json:"title"`
	GoalType        string    `json:"goal_type"`
	Weeks           int       `json:"weeks"`
	WorkoutsPerWeek int       `json:"workouts_per_week"`
	Environment     string    `json:"environment"`
	Status          string    `json:"status"`
	StartDate       time.Time `json:"start_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProgramSession struct {
	ID                  string     `json:"id"`
	ProgramID           string     `json:"program_id"`
	WeekNumber          int        `json:"week_number"`
	DayIndex            int        `json:"day_index"`
	PlannedDate         time.Time  `json:"planned_date"`
	OriginalDate        time.Time  `json:"original_date"`
	Muscle              string     `json:"muscle"`
	SecondaryMuscle     string     `json:"secondary_muscle,omitempty"`
	Environment         string     `json:"environment"`
	Status              string     `json:"status"`
	WorkoutID           *string    `json:"workout_id,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	IsDeload            bool       `json:"is_deload"`
	VolumeMultiplier    float64    `json:"volume_multiplier"`
	IntensityMultiplier float64    `json:"intensity_multiplier"`
	PlannedSets         int        `json:"planned_sets"`
	AdaptationReason    string     `json:"adaptation_reason,omitempty"`
}

type ProgramWithSessions struct {
	Program  Program          `json:"program"`
	Sessions []ProgramSession `json:"sessions"`
}

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id string) (User, error)

	UpsertProfile(ctx context.Context, profile Profile) (Profile, error)
	GetProfile(ctx context.Context, userID string) (Profile, error)
	SetGoal(ctx context.Context, goal Goal) (Goal, error)
	GetGoal(ctx context.Context, userID string) (Goal, error)
	SetTrainingPreferences(ctx context.Context, prefs TrainingPreferences) (TrainingPreferences, error)
	GetTrainingPreferences(ctx context.Context, userID string) (TrainingPreferences, error)
	GetOnboardingStatus(ctx context.Context, userID string) (OnboardingStatus, error)
	CompleteOnboarding(ctx context.Context, userID string) (OnboardingStatus, error)

	SaveRefreshSession(ctx context.Context, session RefreshSession) error
	GetRefreshSession(ctx context.Context, tokenHash string) (RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	CreateWorkout(ctx context.Context, workout Workout, exercises []WorkoutExercise) (WorkoutDetails, error)
	GetWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error)
	StartWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error)
	UpsertWorkoutSet(ctx context.Context, userID, workoutID string, set WorkoutSet) (WorkoutDetails, error)
	ReplaceWorkoutExercise(ctx context.Context, userID, workoutID, workoutExerciseID string, replacement WorkoutExercise) (WorkoutDetails, error)
	CompleteWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error)
	CancelWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error)
	ListWorkouts(ctx context.Context, userID string, limit int) ([]WorkoutDetails, error)
	LastCompletedWorkout(ctx context.Context, userID, muscle, environment string) (WorkoutDetails, error)
	LastExercisePerformance(ctx context.Context, userID, exerciseID string) (ExercisePerformance, error)
	ActiveWorkout(ctx context.Context, userID string) (WorkoutDetails, error)
	SetWorkoutFavorite(ctx context.Context, userID, workoutID string, favorite bool) (WorkoutDetails, error)
	BestPersonalRecord(ctx context.Context, userID, exerciseID, recordType string) (PersonalRecord, error)
	SavePersonalRecord(ctx context.Context, record PersonalRecord) (PersonalRecord, error)
	ListWorkoutPersonalRecords(ctx context.Context, userID, workoutID string) ([]PersonalRecord, error)
	ListPersonalRecords(ctx context.Context, userID string, limit int) ([]PersonalRecord, error)
	UpsertNutritionProfile(ctx context.Context, profile NutritionProfile) (NutritionProfile, error)
	GetNutritionProfile(ctx context.Context, userID string) (NutritionProfile, error)
	SearchFoodItems(ctx context.Context, userID, query string, limit int) ([]FoodItem, error)
	GetFoodItem(ctx context.Context, userID, foodID string) (FoodItem, error)
	GetFoodEntry(ctx context.Context, userID, entryID string) (FoodEntry, error)
	FindFoodByBarcode(ctx context.Context, userID, barcode string) (FoodItem, error)
	CreateCustomFood(ctx context.Context, userID string, food FoodItem) (FoodItem, error)
	CreateFoodEntry(ctx context.Context, entry FoodEntry) (FoodEntry, error)
	ListFoodEntries(ctx context.Context, userID string, from, to time.Time) ([]FoodEntry, error)
	DeleteFoodEntry(ctx context.Context, userID, entryID string) error
	CreateRecipe(ctx context.Context, recipe Recipe) (Recipe, error)
	ListRecipes(ctx context.Context, userID string) ([]Recipe, error)
	GetRecipe(ctx context.Context, userID, recipeID string) (Recipe, error)
	CreateBodyMeasurement(ctx context.Context, measurement BodyMeasurement) (BodyMeasurement, error)
	ListBodyMeasurements(ctx context.Context, userID string, from, to time.Time) ([]BodyMeasurement, error)
	UpsertRecoveryCheckIn(ctx context.Context, checkIn RecoveryCheckIn) (RecoveryCheckIn, error)
	GetRecoveryCheckIn(ctx context.Context, userID, localDate string) (RecoveryCheckIn, error)
	ListRecoveryCheckIns(ctx context.Context, userID string, limit int) ([]RecoveryCheckIn, error)
	UpsertHealthDailySnapshot(ctx context.Context, snapshot HealthDailySnapshot) (HealthDailySnapshot, error)
	GetHealthDailySnapshot(ctx context.Context, userID, localDate string) (HealthDailySnapshot, error)
	ListHealthDailySnapshots(ctx context.Context, userID string, limit int) ([]HealthDailySnapshot, error)
	ListHealthDailySnapshotsForDate(ctx context.Context, userID, localDate string) ([]HealthDailySnapshot, error)
	CreateBodyScan(ctx context.Context, scan BodyScan) (BodyScanDetails, error)
	GetBodyScan(ctx context.Context, userID, scanID string) (BodyScanDetails, error)
	ListBodyScans(ctx context.Context, userID string, limit int) ([]BodyScanDetails, error)
	UpsertBodyScanPhoto(ctx context.Context, photo BodyScanPhoto) (BodyScanDetails, error)
	CompleteBodyScan(ctx context.Context, userID, scanID string, completedAt time.Time) (BodyScanDetails, error)
	DeleteBodyScan(ctx context.Context, userID, scanID string) error
	SaveTechniqueAnalysis(ctx context.Context, analysis TechniqueAnalysis) (TechniqueAnalysis, error)
	UpdateTechniqueAnalysisResult(ctx context.Context, userID, analysisID, resultJSON string) error
	GetTechniqueAnalysis(ctx context.Context, userID, analysisID string) (TechniqueAnalysis, error)
	ListTechniqueAnalyses(ctx context.Context, userID string, limit int) ([]TechniqueAnalysis, error)

	CreateProgram(ctx context.Context, program Program, sessions []ProgramSession) (ProgramWithSessions, error)
	GetProgram(ctx context.Context, userID, programID string) (ProgramWithSessions, error)
	GetActiveProgram(ctx context.Context, userID string) (ProgramWithSessions, error)
	ListPrograms(ctx context.Context, userID string, limit int) ([]ProgramWithSessions, error)
	GetProgramSession(ctx context.Context, userID, sessionID string) (ProgramSession, error)
	UpdateProgramSession(ctx context.Context, userID string, session ProgramSession) (ProgramSession, error)
	FindProgramSessionByWorkout(ctx context.Context, userID, workoutID string) (ProgramSession, error)
	UpdateProgramStatus(ctx context.Context, userID, programID, status string) (ProgramWithSessions, error)
}
