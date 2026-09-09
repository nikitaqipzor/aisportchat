import {resolveApiBaseUrl} from '../config/api';

export type AuthTokens = {
  access_token: string;
  refresh_token: string;
  access_expires_at: string;
};

export type AuthResponse = {
  user: {id: string; email: string; status: string; created_at: string};
  tokens: AuthTokens;
};

export type OnboardingStatus = {
  profile_completed: boolean;
  goal_completed: boolean;
  training_completed: boolean;
  completed: boolean;
};

export type ProfileResponse = {
  profile: {
    height_cm?: number;
    weight_kg?: number;
    experience_level?: 'beginner' | 'intermediate' | 'advanced';
    age_years?: number;
    injuries: string[];
    limitations: string[];
    unit_system: 'metric' | 'imperial';
  };
  goal?: {goal_type: string; target_weight_kg?: number};
  training_preferences?: {
    environments: Array<'home' | 'gym' | 'band'>;
    equipment_ids: string[];
    workouts_per_week: number;
    session_minutes: number;
  };
};

export type WorkoutSet = {
  id: string;
  workout_exercise_id: string;
  set_number: number;
  weight?: number;
  repetitions: number;
  rpe?: number;
  rir?: number;
  completed_at?: string;
};

export type WorkoutExercise = {
  id: string;
  workout_id: string;
  exercise_id: string;
  position: number;
  target_sets: number;
  target_reps_min: number;
  target_reps_max: number;
  target_weight?: number;
  rest_seconds: number;
  progression_note?: string;
};

export type Exercise = {
  id: string;
  name: string;
  primary_muscle: string;
  secondary_muscles?: string[];
  environment: string[];
  equipment: string[];
  difficulty: string;
  movement_pattern: string;
  compound: boolean;
  default_sets: number;
  rep_min: number;
  rep_max: number;
  rest_seconds: number;
  description: string;
  instructions: string[];
  common_mistakes: string[];
  technique_tips: string[];
};

export type ExerciseView = {
  workout_exercise: WorkoutExercise;
  exercise: Exercise;
  sets: WorkoutSet[];
};

export type PersonalRecord = {
  id: string;
  workout_id: string;
  workout_set_id?: string;
  exercise_id: string;
  record_type: 'max_weight' | 'max_reps' | 'estimated_1rm';
  value: number;
  previous_value?: number;
  achieved_at: string;
};

export type MuscleStats = {
  muscle: string;
  completed_workouts: number;
  total_sets: number;
  total_volume: number;
  recent_volume: number;
  last_workout_volume: number;
  last_workout_at?: string;
  days_since_last_workout?: number;
  personal_records_count: number;
};

export type HistoryFilters = {
  limit?: number;
  muscle?: string;
  environment?: 'home' | 'gym' | 'band';
  status?: 'planned' | 'active' | 'completed' | 'cancelled';
  favorite?: boolean;
};

export type WorkoutProgressSummary = {
  period_days: number; completed_workouts: number; training_days: number; workouts_per_week: number;
  total_sets: number; total_volume: number; recent_volume_7d: number; previous_volume_7d: number;
  weekly_streak: number; personal_record_count: number;
  exercise_bests: Array<{exercise_id: string; exercise_name: string; max_weight?: number; max_reps: number}>;
};

export type WorkoutView = {
  workout: {
    id: string;
    muscle: string;
    environment: 'home' | 'gym' | 'band';
    status: 'planned' | 'active' | 'completed' | 'cancelled';
    duration_minutes: number;
    started_at?: string;
    completed_at?: string;
    duration_seconds?: number;
    total_volume: number;
    completion_percent: number;
    ended_early: boolean;
    favorite: boolean;
    created_at: string;
  };
  exercises: ExerciseView[];
  personal_records: PersonalRecord[];
};

export type ManualWorkoutExerciseInput = {
  exercise_id: string;
  target_sets: number;
  target_reps: number;
  target_weight?: number;
  rest_seconds: number;
};

export type FinishResult = {
  workout: WorkoutView;
  personal_records: PersonalRecord[];
  next_recommendations: Array<{
    exercise_id: string;
    exercise_name: string;
    current_weight?: number;
    next_weight?: number;
    message: string;
  }>;
};


export type ProgramSession = {
  id: string;
  program_id: string;
  week_number: number;
  day_index: number;
  planned_date: string;
  original_date: string;
  muscle: string;
  secondary_muscle?: string;
  environment: 'home' | 'gym' | 'band';
  status: 'planned' | 'rescheduled' | 'missed' | 'completed' | 'skipped';
  workout_id?: string;
  completed_at?: string;
  is_deload: boolean;
  volume_multiplier: number;
  intensity_multiplier: number;
  planned_sets: number;
  adaptation_reason?: string;
};

export type TrainingProgram = {
  program: {
    id: string;
    title: string;
    goal_type: string;
    weeks: 4 | 8 | 12;
    workouts_per_week: number;
    environment: 'home' | 'gym' | 'band';
    status: 'active' | 'completed' | 'archived';
    start_date: string;
    created_at: string;
    updated_at: string;
  };
  sessions: ProgramSession[];
};

export type ProgramAnalytics = {
  program_id: string;
  planned_sessions: number;
  completed_sessions: number;
  missed_sessions: number;
  rescheduled_sessions: number;
  upcoming_sessions: number;
  deload_sessions: number;
  adherence_percent: number;
  current_week: number;
  total_weeks: number;
  planned_sets_by_muscle: Record<string, number>;
  completed_sets_by_muscle: Record<string, number>;
  next_session?: ProgramSession;
};

export type ProgramWorkoutResponse = {session: ProgramSession; workout: WorkoutView};
export type ProgramExplanation = {
  session: ProgramSession;
  explanation: {message: string; model: string; provider: string; tool_calls?: string[]};
};

export type NutritionProfile = {
  user_id: string;
  goal: 'lose' | 'recomp' | 'maintain' | 'gain' | 'strength' | 'endurance';
  activity_level: 'low' | 'light' | 'moderate' | 'high' | 'athlete';
  calculation_mode: 'auto' | 'manual';
  calorie_target: number;
  protein_target_g: number;
  fat_target_g: number;
  carb_target_g: number;
  updated_at: string;
};

export type FoodItem = {
  id: string;
  name: string;
  brand?: string;
  barcode?: string;
  kcal_per_100g: number;
  protein_per_100g: number;
  fat_per_100g: number;
  carbs_per_100g: number;
  fiber_per_100g: number;
  serving_g: number;
  source: string;
};


export type RecipeItem = {
  food_id: string;
  food_name: string;
  quantity_g: number;
  calories: number;
  protein_g: number;
  fat_g: number;
  carbs_g: number;
};

export type Recipe = {
  id: string;
  name: string;
  items: RecipeItem[];
  created_at: string;
  updated_at: string;
};

export type BodyMeasurement = {
  id: string;
  logged_at: string;
  weight_kg?: number;
  waist_cm?: number;
  chest_cm?: number;
  shoulder_cm?: number;
  arm_cm?: number;
  forearm_cm?: number;
  hip_cm?: number;
  thigh_cm?: number;
  calf_cm?: number;
  neck_cm?: number;
};

export type BodyScanPhoto = {
  id: string;
  scan_id: string;
  view: 'front' | 'side' | 'back';
  mime_type: string;
  width: number;
  height: number;
  bytes: number;
  brightness: number;
  contrast: number;
  quality_status: 'accepted' | 'warning';
  quality_issues?: string[];
  created_at: string;
};

export type BodyScanDetails = {
  scan: {id: string; status: 'draft' | 'completed'; created_at: string; completed_at?: string};
  photos: BodyScanPhoto[];
};

export type BodyScanComparison = {
  from_scan_id: string;
  to_scan_id: string;
  days_between: number;
  capture_consistency_score: number;
  lighting_delta: number;
  warnings?: string[];
  weight_delta_kg?: number;
  waist_delta_cm?: number;
  visual_cv_status: string;
};

export type ProgressSummary = {
  days: number;
  points: number;
  latest?: BodyMeasurement;
  weight_start_kg?: number;
  weight_current_kg?: number;
  weight_delta_kg?: number;
  waist_start_cm?: number;
  waist_current_cm?: number;
  waist_delta_cm?: number;
};

export type NutritionCorrelation = {
  days: number;
  logged_training_days: number;
  logged_rest_days: number;
  avg_training_calories: number;
  avg_rest_calories: number;
  avg_training_protein_g: number;
  avg_rest_protein_g: number;
  training_calorie_adherence_percent: number;
  rest_calorie_adherence_percent: number;
};

export type FoodEntry = {
  id: string;
  food_id: string;
  food_name: string;
  meal_type: 'breakfast' | 'lunch' | 'dinner' | 'snack';
  logged_at: string;
  quantity_g: number;
  calories: number;
  protein_g: number;
  fat_g: number;
  carbs_g: number;
  fiber_g: number;
};

export type NutritionDay = {
  date: string;
  profile: NutritionProfile;
  consumed_calories: number;
  consumed_protein_g: number;
  consumed_fat_g: number;
  consumed_carbs_g: number;
  consumed_fiber_g: number;
  remaining_calories: number;
  remaining_protein_g: number;
  remaining_fat_g: number;
  remaining_carbs_g: number;
  entries: FoodEntry[];
  training_day: boolean;
  completed_workouts: number;
};

export type NutritionHistoryItem = {
  date: string;
  calories: number;
  protein_g: number;
  fat_g: number;
  carbs_g: number;
  target_calories: number;
  adherence_percent: number;
  training_day: boolean;
};



export type HealthDailySnapshot = {
  id: string;
  date: string;
  provider: 'health_connect';
  source_package: string;
  source_label: string;
  steps: number;
  distance_m: number;
  active_calories_kcal: number;
  sleep_minutes: number;
  deep_sleep_minutes: number;
  light_sleep_minutes: number;
  rem_sleep_minutes: number;
  awake_minutes: number;
  exercise_minutes: number;
  exercise_sessions: number;
  exercise_heart_rate_avg?: number;
  exercise_heart_rate_max?: number;
  resting_heart_rate?: number;
  data_types: string[];
  captured_at: string;
  imported_at: string;
};

export type HealthBaselineMetric = {
  average: number;
  sample_days: number;
  unit: string;
};

export type HealthBaselineWindow = {
  window_days: number;
  available_days: number;
  coverage_percent: number;
  sleep_minutes?: HealthBaselineMetric;
  steps?: HealthBaselineMetric;
  active_calories_kcal?: HealthBaselineMetric;
  exercise_minutes?: HealthBaselineMetric;
  resting_heart_rate?: HealthBaselineMetric;
};

export type HealthMetricDeviation = {
  current: number;
  baseline: number;
  delta: number;
  delta_percent: number;
  unit: string;
};

export type HealthInsights = {
  date: string;
  snapshot?: HealthDailySnapshot;
  baseline_7d: HealthBaselineWindow;
  baseline_28d: HealthBaselineWindow;
  deviations: Record<string, HealthMetricDeviation>;
  freshness: {status: 'fresh' | 'aging' | 'stale' | 'missing'; sync_age_minutes: number; captured_at?: string; imported_at?: string};
  provenance: {metric: string; available: boolean; source_package?: string; source_label?: string}[];
  sources: {source_package: string; source_label: string; selected: boolean; imported_at: string; data_types: string[]}[];
  conflict_resolved: boolean;
  confidence: 'low' | 'medium' | 'high';
  confidence_percent: number;
  reasons: string[];
};

export type RecoveryCheckIn = {
  id: string;
  date: string;
  sleep_hours: number;
  sleep_quality: number;
  energy: number;
  stress: number;
  muscle_soreness: Record<string, number>;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type RecoveryMuscleStatus = {
  muscle: string;
  score: number;
  status: 'ready' | 'moderate' | 'fatigued';
  soreness: number;
  recent_sets_7d: number;
  recent_sets_48h: number;
  hours_since_last_workout?: number;
};

export type Readiness = {
  date: string;
  check_in_completed: boolean;
  score: number;
  status: 'ready' | 'good' | 'moderate' | 'low';
  factors: {sleep: number; energy: number; stress: number; soreness: number; training_load: number};
  volume_multiplier: number;
  intensity_multiplier: number;
  reasons: string[];
  muscles: RecoveryMuscleStatus[];
  check_in?: RecoveryCheckIn;
  wearable?: HealthDailySnapshot;
  sleep_source?: 'none' | 'check_in' | 'wearable';
  health_insights?: HealthInsights;
};

export type ReadinessFactorImpact = {
  key: 'sleep' | 'energy' | 'stress' | 'soreness' | 'training_load';
  label: string;
  today_score: number;
  previous_score?: number;
  weight: number;
  weighted_points: number;
  weighted_delta_points?: number;
  direction: 'up' | 'down' | 'stable';
  message: string;
};

export type ReadinessHealthTrend = {
  key: string;
  label: string;
  unit: string;
  current?: number;
  previous?: number;
  baseline_7d?: number;
  baseline_28d?: number;
  delta_vs_7d?: number;
  delta_vs_28d?: number;
  affects_readiness: boolean;
};

export type ReadinessTimeline = {
  date: string;
  score: number;
  status: Readiness['status'];
  previous_date: string;
  previous_score?: number;
  score_delta?: number;
  comparison_available: boolean;
  factors: ReadinessFactorImpact[];
  health_trends: ReadinessHealthTrend[];
  adaptation: {
    volume_multiplier: number;
    intensity_multiplier: number;
    volume_change_percent: number;
    intensity_change_percent: number;
  };
  summary: string[];
};

export type AIFoodDraftItem = {
  name: string;
  quantity_g: number;
  confidence: number;
  notes?: string;
  match_quality: string;
  matched_food?: FoodItem;
  estimated_calories?: number;
  estimated_protein_g?: number;
  estimated_fat_g?: number;
  estimated_carbs_g?: number;
};

export type AIFoodDraft = {
  source: 'text' | 'image';
  text?: string;
  items: AIFoodDraftItem[];
  warning?: string;
};

export type AIChatMessage = {role: 'user' | 'assistant'; content: string};
export type AIChatResponse = {message: string; model: string; provider: string; tool_calls?: string[]};
export type AIStatus = {provider: string; model: string};
export type WeeklyAIReport = {
  stats: {
    from_date: string; to_date: string; completed_workouts: number; training_volume: number;
    logged_nutrition_days: number; avg_calories: number; avg_protein_g: number; calorie_target?: number; protein_target_g?: number;
    weight_start_kg?: number; weight_current_kg?: number; weight_delta_kg?: number; new_prs: number; recovery_checkin_days?: number; avg_readiness?: number;
  };
  summary: string; wins: string[]; focus: string[]; next_actions: string[]; model: string; provider: string;
};

const API_BASE_URL = resolveApiBaseUrl();

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

type AuthSessionAdapter = {
  loadTokens: () => Promise<AuthTokens | null>;
  saveTokens: (tokens: AuthTokens) => Promise<void>;
  clearTokens?: () => Promise<void>;
  onTokensChanged?: (tokens: AuthTokens | null) => void;
};

let authSessionAdapter: AuthSessionAdapter | null = null;
let latestTokens: AuthTokens | null = null;
let refreshInFlight: Promise<AuthTokens> | null = null;

async function parseResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new ApiError(response.status, body.error ?? `Request failed: ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

async function rawRequest<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? {Authorization: `Bearer ${accessToken}`} : {}),
      ...(options.headers ?? {}),
    },
  });
  return parseResponse<T>(response);
}

async function getCurrentTokens(): Promise<AuthTokens | null> {
  if (latestTokens) return latestTokens;
  if (!authSessionAdapter) return null;
  latestTokens = await authSessionAdapter.loadTokens();
  return latestTokens;
}

async function rotateAccessToken(fallbackRefreshToken?: string): Promise<AuthTokens> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    const current = await getCurrentTokens();
    const refreshToken = current?.refresh_token ?? fallbackRefreshToken;
    if (!refreshToken) throw new ApiError(401, 'Session expired');
    try {
      const response = await rawRequest<{tokens: AuthTokens}>('/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({refresh_token: refreshToken}),
      });
      latestTokens = response.tokens;
      await authSessionAdapter?.saveTokens(response.tokens);
      authSessionAdapter?.onTokensChanged?.(response.tokens);
      return response.tokens;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        latestTokens = null;
        await authSessionAdapter?.clearTokens?.();
        authSessionAdapter?.onTokensChanged?.(null);
      }
      throw error;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

async function request<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  if (!accessToken) return rawRequest<T>(path, options);

  const current = await getCurrentTokens();
  const effectiveAccess = current?.access_token ?? accessToken;
  try {
    return await rawRequest<T>(path, options, effectiveAccess);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401 || path.startsWith('/auth/')) throw error;
    const fresh = await rotateAccessToken(current?.refresh_token);
    return rawRequest<T>(path, options, fresh.access_token);
  }
}


export type PoseLandmarkPayload = {x: number; y: number; z: number; visibility: number};
export type PoseFramePayload = {timestamp_ms: number; landmarks: PoseLandmarkPayload[]};
export type TechniqueRepMetric = {number: number; start_ms: number; end_ms: number; duration_ms: number; eccentric_ms: number; concentric_ms: number; rom_degrees: number; left_rom_degrees?: number; right_rom_degrees?: number; symmetry_delta_degrees?: number};
export type TechniqueResult = {
  id: string; exercise_key: string; exercise_name: string; capture_mode: 'recorded'|'live'; workout_id?: string; workout_exercise_id?: string; set_number?: number;
  rep_count: number; technique_score: number; rom_score: number; tempo_score: number; symmetry_score: number; stability_score: number; confidence: number; duration_ms: number; average_eccentric_ms: number; average_concentric_ms: number; average_rep_rpm: number; algorithm_version: string; reps: TechniqueRepMetric[]; feedback: string[]; created_at: string;
};
export type TechniqueAnalyzePayload = {
  exercise_key: string; capture_mode?: 'recorded'|'live'; workout_id?: string; workout_exercise_id?: string; set_number?: number; duration_ms: number; frames: PoseFramePayload[];
};
export type TechniqueExercise = {key: 'squat'|'biceps_curl'|'push_up'|'lunge'|'shoulder_press'; name: string};
export const api = {
  configureAuthSession(adapter: AuthSessionAdapter | null) {
    authSessionAdapter = adapter;
    if (!adapter) { latestTokens = null; refreshInFlight = null; }
  },
  setCurrentTokens(tokens: AuthTokens | null) {
    latestTokens = tokens;
  },
  apiBaseUrl() {
    return API_BASE_URL;
  },
  isUnauthorized(error: unknown) {
    return error instanceof ApiError && error.status === 401;
  },
  isNetworkError(error: unknown) {
    return !(error instanceof ApiError);
  },
  register(email: string, password: string) {
    return request<AuthResponse>('/auth/register', {method: 'POST', body: JSON.stringify({email, password})});
  },
  login(email: string, password: string) {
    return request<AuthResponse>('/auth/login', {method: 'POST', body: JSON.stringify({email, password})});
  },
  refresh(refreshToken: string) {
    return request<{tokens: AuthTokens}>('/auth/refresh', {method: 'POST', body: JSON.stringify({refresh_token: refreshToken})});
  },
  logout(refreshToken: string) {
    return request<void>('/auth/logout', {method: 'POST', body: JSON.stringify({refresh_token: refreshToken})});
  },
  onboardingStatus(accessToken: string) {
    return request<OnboardingStatus>('/onboarding/status', {method: 'GET'}, accessToken);
  },
  getProfile(accessToken: string) {
    return request<ProfileResponse>('/profile', {method: 'GET'}, accessToken);
  },
  updateProfile(accessToken: string, payload: {height_cm: number; weight_kg: number; experience_level: string; age_years?: number; injuries?: string[]; limitations?: string[]; unit_system: 'metric'}) {
    return request('/profile', {method: 'PATCH', body: JSON.stringify(payload)}, accessToken);
  },
  setGoal(accessToken: string, goalType: string, targetWeightKG?: number) {
    return request('/profile/goal', {method: 'PUT', body: JSON.stringify({goal_type: goalType, target_weight_kg: targetWeightKG})}, accessToken);
  },
  setTrainingPreferences(accessToken: string, payload: {environments: string[]; equipment_ids: string[]; workouts_per_week: number; session_minutes: number}) {
    return request('/profile/training-preferences', {method: 'PUT', body: JSON.stringify(payload)}, accessToken);
  },
  completeOnboarding(accessToken: string) {
    return request<OnboardingStatus>('/onboarding/complete', {method: 'POST', body: '{}'}, accessToken);
  },
  generateWorkout(accessToken: string, muscle: string, environment: 'home' | 'gym' | 'band', localDate?: string) {
    return request<WorkoutView>('/workouts/generate', {method: 'POST', body: JSON.stringify({muscle, environment, local_date: localDate})}, accessToken);
  },
  listExercises(accessToken: string, filters: {muscle?: string; environment?: 'home'|'gym'|'band'; query?: string} = {}) {
    const query = new URLSearchParams();
    if (filters.muscle) query.set('muscle', filters.muscle);
    if (filters.environment) query.set('environment', filters.environment);
    if (filters.query) query.set('query', filters.query);
    return request<{items: Exercise[]}>(`/exercises?${query.toString()}`, {method: 'GET'}, accessToken);
  },
  createManualWorkout(accessToken: string, payload: {muscle: string; environment: 'home'|'gym'|'band'; duration_minutes: number; exercises: ManualWorkoutExerciseInput[]}) {
    return request<WorkoutView>('/workouts/manual', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  activeWorkout(accessToken: string) {
    return request<WorkoutView | undefined>('/workouts/active', {method: 'GET'}, accessToken);
  },
  getWorkout(accessToken: string, workoutId: string) {
    return request<WorkoutView>(`/workouts/${workoutId}`, {method: 'GET'}, accessToken);
  },
  startWorkout(accessToken: string, workoutId: string) {
    return request<WorkoutView>(`/workouts/${workoutId}/start`, {method: 'POST', body: '{}'}, accessToken);
  },
  logSet(accessToken: string, workoutId: string, payload: {workout_exercise_id: string; set_number: number; weight?: number; repetitions: number; rpe?: number; rir?: number}) {
    return request<WorkoutView>(`/workouts/${workoutId}/sets`, {method: 'PUT', body: JSON.stringify(payload)}, accessToken);
  },
  replaceExercise(accessToken: string, workoutId: string, workoutExerciseId: string, reason = 'user_request') {
    return request<WorkoutView>(`/workouts/${workoutId}/exercises/${workoutExerciseId}/replace`, {method: 'POST', body: JSON.stringify({reason})}, accessToken);
  },
  finishWorkout(accessToken: string, workoutId: string) {
    return request<FinishResult>(`/workouts/${workoutId}/finish`, {method: 'POST', body: '{}'}, accessToken);
  },
  cancelWorkout(accessToken: string, workoutId: string) {
    return request<WorkoutView>(`/workouts/${workoutId}/cancel`, {method: 'POST', body: '{}'}, accessToken);
  },
  workoutHistory(accessToken: string, filters: HistoryFilters = {}) {
    const query = new URLSearchParams();
    query.set('limit', String(filters.limit ?? 30));
    if (filters.muscle) query.set('muscle', filters.muscle);
    if (filters.environment) query.set('environment', filters.environment);
    if (filters.status) query.set('status', filters.status);
    if (filters.favorite !== undefined) query.set('favorite', String(filters.favorite));
    return request<{items: WorkoutView[]}>(`/workouts/history?${query.toString()}`, {method: 'GET'}, accessToken);
  },
  workoutProgressSummary(accessToken: string, days = 28) {
    return request<WorkoutProgressSummary>(`/workouts/progress-summary?days=${days}`, {method: 'GET'}, accessToken);
  },
  muscleStats(accessToken: string, muscle: string) {
    return request<MuscleStats>(`/muscles/${muscle}/stats`, {method: 'GET'}, accessToken);
  },
  personalRecords(accessToken: string, limit = 50) {
    return request<{items: PersonalRecord[]}>(`/records?limit=${limit}`, {method: 'GET'}, accessToken);
  },
  setWorkoutFavorite(accessToken: string, workoutId: string, favorite: boolean) {
    return request<WorkoutView>(`/workouts/${workoutId}/favorite`, {method: 'PATCH', body: JSON.stringify({favorite})}, accessToken);
  },
  repeatWorkout(accessToken: string, workoutId: string) {
    return request<WorkoutView>(`/workouts/${workoutId}/repeat`, {method: 'POST', body: '{}'}, accessToken);
  },
  generateProgram(accessToken: string, payload: {weeks: 4 | 8 | 12; workouts_per_week?: number; environment?: 'home' | 'gym' | 'band'; start_date?: string}) {
    return request<TrainingProgram>('/programs/generate', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  activeProgram(accessToken: string) {
    return request<TrainingProgram | undefined>('/programs/active', {method: 'GET'}, accessToken);
  },
  programHistory(accessToken: string, limit = 20) {
    return request<{items: TrainingProgram[]}>(`/programs/history?limit=${limit}`, {method: 'GET'}, accessToken);
  },
  getProgram(accessToken: string, programId: string) {
    return request<TrainingProgram>(`/programs/${programId}`, {method: 'GET'}, accessToken);
  },
  programAnalytics(accessToken: string, programId: string) {
    return request<ProgramAnalytics>(`/programs/${programId}/analytics`, {method: 'GET'}, accessToken);
  },
  archiveProgram(accessToken: string, programId: string) {
    return request<TrainingProgram>(`/programs/${programId}/archive`, {method: 'POST', body: '{}'}, accessToken);
  },
  createProgramWorkout(accessToken: string, sessionId: string, localDate?: string) {
    const query = localDate ? `?date=${encodeURIComponent(localDate)}` : '';
    return request<ProgramWorkoutResponse>(`/programs/sessions/${sessionId}/workout${query}`, {method: 'POST', body: '{}'}, accessToken);
  },
  rescheduleProgramSession(accessToken: string, sessionId: string, newDate: string, reason?: string) {
    return request<ProgramSession>(`/programs/sessions/${sessionId}/reschedule`, {method: 'POST', body: JSON.stringify({new_date: newDate, reason})}, accessToken);
  },
  autoRescheduleProgramSession(accessToken: string, sessionId: string) {
    return request<ProgramSession>(`/programs/sessions/${sessionId}/auto-reschedule`, {method: 'POST', body: '{}'}, accessToken);
  },
  explainProgramSession(accessToken: string, sessionId: string) {
    return request<ProgramExplanation>(`/programs/sessions/${sessionId}/explain`, {method: 'POST', body: '{}'}, accessToken);
  },

  recoveryTimeline(accessToken: string, date?: string) {
    const query = date ? `?date=${encodeURIComponent(date)}` : '';
    return request<ReadinessTimeline>(`/recovery/timeline${query}`, {method: 'GET'}, accessToken);
  },
  recoveryToday(accessToken: string, date?: string) {
    const query = date ? `?date=${encodeURIComponent(date)}` : '';
    return request<Readiness>(`/recovery/today${query}`, {method: 'GET'}, accessToken);
  },
  saveRecoveryCheckIn(accessToken: string, payload: {date: string; sleep_hours: number; sleep_quality: number; energy: number; stress: number; muscle_soreness: Record<string, number>; notes?: string}) {
    return request<Readiness>('/recovery/check-in', {method: 'PUT', body: JSON.stringify(payload)}, accessToken);
  },
  recoveryHistory(accessToken: string, limit = 30) {
    return request<{items: RecoveryCheckIn[]}>(`/recovery/history?limit=${limit}`, {method: 'GET'}, accessToken);
  },

  importHealthSnapshot(accessToken: string, payload: Omit<HealthDailySnapshot, 'id' | 'imported_at'>) {
    return request<HealthDailySnapshot>('/health/snapshots', {method: 'PUT', body: JSON.stringify(payload)}, accessToken);
  },
  healthToday(accessToken: string, date?: string) {
    const query = date ? `?date=${encodeURIComponent(date)}` : '';
    return request<HealthDailySnapshot | undefined>(`/health/today${query}`, {method: 'GET'}, accessToken);
  },
  healthHistory(accessToken: string, limit = 14) {
    return request<{items: HealthDailySnapshot[]}>(`/health/history?limit=${limit}`, {method: 'GET'}, accessToken);
  },
  healthInsights(accessToken: string, date?: string) {
    const query = date ? `?date=${encodeURIComponent(date)}` : '';
    return request<HealthInsights>(`/health/insights${query}`, {method: 'GET'}, accessToken);
  },

  nutritionProfile(accessToken: string) {
    return request<NutritionProfile | undefined>('/nutrition/profile', {method: 'GET'}, accessToken);
  },
  setNutritionProfile(accessToken: string, payload: {
    goal: NutritionProfile['goal'];
    activity_level: NutritionProfile['activity_level'];
    calculation_mode: 'auto' | 'manual';
    calorie_target?: number;
    protein_target_g?: number;
    fat_target_g?: number;
    carb_target_g?: number;
  }) {
    return request<NutritionProfile>('/nutrition/profile', {method: 'PUT', body: JSON.stringify(payload)}, accessToken);
  },
  nutritionToday(accessToken: string, date?: string) {
    const query = date ? `?date=${encodeURIComponent(date)}` : '';
    return request<NutritionDay>(`/nutrition/today${query}`, {method: 'GET'}, accessToken);
  },
  nutritionHistory(accessToken: string, days = 7) {
    return request<{items: NutritionHistoryItem[]}>(`/nutrition/history?days=${days}`, {method: 'GET'}, accessToken);
  },
  searchFoods(accessToken: string, query: string, limit = 30) {
    return request<{items: FoodItem[]}>(`/nutrition/foods?query=${encodeURIComponent(query)}&limit=${limit}`, {method: 'GET'}, accessToken);
  },
  logFood(accessToken: string, payload: {food_id: string; meal_type: FoodEntry['meal_type']; quantity_g: number; logged_at?: string}) {
    return request<NutritionDay>('/nutrition/entries', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  deleteFoodEntry(accessToken: string, entryId: string) {
    return request<void>(`/nutrition/entries/${entryId}`, {method: 'DELETE'}, accessToken);
  },

  createCustomFood(accessToken: string, payload: {name: string; brand?: string; barcode?: string; kcal_per_100g: number; protein_per_100g: number; fat_per_100g: number; carbs_per_100g: number; fiber_per_100g: number; serving_g: number}) {
    return request<FoodItem>('/nutrition/foods/custom', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  foodByBarcode(accessToken: string, barcode: string) {
    return request<FoodItem>(`/nutrition/foods/barcode/${encodeURIComponent(barcode)}`, {method: 'GET'}, accessToken);
  },
  repeatFoodEntry(accessToken: string, entryId: string, meal_type?: FoodEntry['meal_type']) {
    return request<NutritionDay>(`/nutrition/entries/${entryId}/repeat`, {method: 'POST', body: JSON.stringify({meal_type})}, accessToken);
  },
  recipes(accessToken: string) {
    return request<{items: Recipe[]}>('/nutrition/recipes', {method: 'GET'}, accessToken);
  },
  createRecipe(accessToken: string, payload: {name: string; items: Array<{food_id: string; quantity_g: number}>}) {
    return request<Recipe>('/nutrition/recipes', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  logRecipe(accessToken: string, recipeId: string, payload: {meal_type: FoodEntry['meal_type']; scale?: number}) {
    return request<NutritionDay>(`/nutrition/recipes/${recipeId}/log`, {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  nutritionCorrelation(accessToken: string, days = 30) {
    return request<NutritionCorrelation>(`/nutrition/correlation?days=${days}`, {method: 'GET'}, accessToken);
  },
  logMeasurement(accessToken: string, payload: Omit<BodyMeasurement, 'id' | 'logged_at'> & {logged_at?: string}) {
    return request<BodyMeasurement>('/progress/measurements', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  measurementHistory(accessToken: string, days = 30) {
    return request<{items: BodyMeasurement[]}>(`/progress/measurements?days=${days}`, {method: 'GET'}, accessToken);
  },
  progressSummary(accessToken: string, days = 30) {
    return request<ProgressSummary>(`/progress/summary?days=${days}`, {method: 'GET'}, accessToken);
  },
  createBodyScan(accessToken: string) {
    return request<BodyScanDetails>('/body-scans', {method: 'POST', body: '{}'}, accessToken);
  },
  bodyScans(accessToken: string, limit = 20) {
    return request<{items: BodyScanDetails[]}>(`/body-scans?limit=${limit}`, {method: 'GET'}, accessToken);
  },
  bodyScan(accessToken: string, scanId: string) {
    return request<BodyScanDetails>(`/body-scans/${scanId}`, {method: 'GET'}, accessToken);
  },
  deleteBodyScan(accessToken: string, scanId: string) {
    return request<void>(`/body-scans/${scanId}`, {method: 'DELETE'}, accessToken);
  },
  putBodyScanPhoto(accessToken: string, scanId: string, view: 'front'|'side'|'back', imageDataUrl: string) {
    return request<BodyScanDetails>(`/body-scans/${scanId}/photos/${view}`, {method: 'PUT', body: JSON.stringify({image_data_url: imageDataUrl})}, accessToken);
  },
  completeBodyScan(accessToken: string, scanId: string) {
    return request<BodyScanDetails>(`/body-scans/${scanId}/complete`, {method: 'POST', body: '{}'}, accessToken);
  },
  latestBodyScanComparison(accessToken: string) {
    return request<BodyScanComparison>('/body-scans/comparison/latest', {method: 'GET'}, accessToken);
  },
  techniqueExercises(accessToken: string) {
    return request<{items: TechniqueExercise[]; algorithm_version: string}>('/technique/exercises', {method: 'GET'}, accessToken);
  },
  analyzeTechnique(accessToken: string, payload: TechniqueAnalyzePayload) {
    return request<TechniqueResult>('/technique/analyses', {method: 'POST', body: JSON.stringify(payload)}, accessToken);
  },
  techniqueHistory(accessToken: string, limit = 20) {
    return request<{items: TechniqueResult[]}>(`/technique/analyses?limit=${limit}`, {method: 'GET'}, accessToken);
  },
  techniqueAnalysis(accessToken: string, analysisId: string) {
    return request<TechniqueResult>(`/technique/analyses/${analysisId}`, {method: 'GET'}, accessToken);
  },
  aiStatus(accessToken: string) {
    return request<AIStatus>('/ai/status', {method: 'GET'}, accessToken);
  },
  aiParseFoodText(accessToken: string, text: string) {
    return request<AIFoodDraft>('/ai/food/parse', {method: 'POST', body: JSON.stringify({text})}, accessToken);
  },
  aiParseFoodPhoto(accessToken: string, imageDataUrl: string) {
    return request<AIFoodDraft>('/ai/food/photo', {method: 'POST', body: JSON.stringify({image_data_url: imageDataUrl})}, accessToken);
  },
  aiConfirmFood(accessToken: string, mealType: 'breakfast'|'lunch'|'dinner'|'snack', items: Array<{food_id: string; quantity_g: number}>) {
    return request<NutritionDay>('/ai/food/confirm', {method: 'POST', body: JSON.stringify({meal_type: mealType, items})}, accessToken);
  },
  aiChat(accessToken: string, message: string, history: AIChatMessage[] = [], signal?: AbortSignal) {
    return request<AIChatResponse>('/ai/chat', {method: 'POST', body: JSON.stringify({message, history}), signal}, accessToken);
  },
  aiWeeklyReport(accessToken: string) {
    return request<WeeklyAIReport>('/ai/reports/weekly', {method: 'GET'}, accessToken);
  },

};
