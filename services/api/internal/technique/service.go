package technique

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

const AlgorithmVersion = "technique-v1.1"

type Landmark struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
	Visibility float64 `json:"visibility"`
}
type Frame struct {
	TimestampMS int64      `json:"timestamp_ms"`
	Landmarks   []Landmark `json:"landmarks"`
}
type AnalyzeInput struct {
	ExerciseKey       string  `json:"exercise_key"`
	CaptureMode       string  `json:"capture_mode,omitempty"`
	WorkoutID         string  `json:"workout_id,omitempty"`
	WorkoutExerciseID string  `json:"workout_exercise_id,omitempty"`
	SetNumber         int     `json:"set_number,omitempty"`
	DurationMS        int64   `json:"duration_ms"`
	Frames            []Frame `json:"frames"`
}
type RepMetric struct {
	Number          int     `json:"number"`
	StartMS         int64   `json:"start_ms"`
	EndMS           int64   `json:"end_ms"`
	DurationMS      int64   `json:"duration_ms"`
	EccentricMS     int64   `json:"eccentric_ms"`
	ConcentricMS    int64   `json:"concentric_ms"`
	ROMDegrees      float64 `json:"rom_degrees"`
	LeftROMDegrees  float64 `json:"left_rom_degrees,omitempty"`
	RightROMDegrees float64 `json:"right_rom_degrees,omitempty"`
	SymmetryDelta   float64 `json:"symmetry_delta_degrees,omitempty"`
}
type Result struct {
	ID                  string      `json:"id"`
	ExerciseKey         string      `json:"exercise_key"`
	ExerciseName        string      `json:"exercise_name"`
	CaptureMode         string      `json:"capture_mode"`
	WorkoutID           string      `json:"workout_id,omitempty"`
	WorkoutExerciseID   string      `json:"workout_exercise_id,omitempty"`
	SetNumber           int         `json:"set_number,omitempty"`
	RepCount            int         `json:"rep_count"`
	TechniqueScore      int         `json:"technique_score"`
	ROMScore            int         `json:"rom_score"`
	TempoScore          int         `json:"tempo_score"`
	SymmetryScore       int         `json:"symmetry_score"`
	StabilityScore      int         `json:"stability_score"`
	Confidence          float64     `json:"confidence"`
	DurationMS          int64       `json:"duration_ms"`
	AverageEccentricMS  int64       `json:"average_eccentric_ms"`
	AverageConcentricMS int64       `json:"average_concentric_ms"`
	AverageRepRPM       float64     `json:"average_rep_rpm"`
	AlgorithmVersion    string      `json:"algorithm_version"`
	Reps                []RepMetric `json:"reps"`
	Feedback            []string    `json:"feedback"`
	CreatedAt           time.Time   `json:"created_at"`
}

type Service struct{ st store.Store }

func NewService(st store.Store) *Service { return &Service{st: st} }

type exerciseRule struct {
	Name               string
	Bottom, Top        float64
	MinROM             float64
	MinRepMS, MaxRepMS int64
	StartAt            string
	FirstPhase         string
	Angle              func(Frame) (float64, float64, bool)
	Stability          func(Frame) float64
}

var rules = map[string]exerciseRule{
	"squat":          {"Приседания", 105, 155, 45, 900, 6000, "top", "eccentric", kneeAngles, trunkStability},
	"biceps_curl":    {"Подъём на бицепс", 75, 145, 55, 650, 5000, "top", "concentric", elbowAngles, shoulderStability},
	"push_up":        {"Отжимания", 105, 155, 40, 700, 5000, "top", "eccentric", elbowAngles, plankStability},
	"lunge":          {"Выпады", 110, 150, 35, 900, 6500, "top", "eccentric", lungeAngles, trunkStability},
	"shoulder_press": {"Жим над головой", 105, 150, 35, 700, 5000, "bottom", "concentric", elbowAngles, shoulderStability},
}

func SupportedExercises() []map[string]string {
	keys := make([]string, 0, len(rules))
	for k := range rules {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]map[string]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, map[string]string{"key": k, "name": rules[k].Name})
	}
	return out
}

func (s *Service) Analyze(ctx context.Context, userID string, in AnalyzeInput) (Result, error) {
	key := strings.TrimSpace(in.ExerciseKey)
	rule, ok := rules[key]
	if !ok {
		return Result{}, errors.New("unsupported exercise")
	}
	mode := strings.TrimSpace(in.CaptureMode)
	if mode == "" {
		mode = "recorded"
	}
	if mode != "recorded" && mode != "live" {
		return Result{}, errors.New("capture_mode must be recorded or live")
	}
	if err := s.validateWorkoutLink(ctx, userID, key, in); err != nil {
		return Result{}, err
	}
	if len(in.Frames) < 8 || len(in.Frames) > 1200 {
		return Result{}, errors.New("pose frames must contain 8..1200 frames")
	}
	last := int64(-1)
	valid := 0
	confSum := 0.0
	for _, f := range in.Frames {
		if f.TimestampMS <= last || len(f.Landmarks) < 33 {
			return Result{}, errors.New("invalid pose frame sequence")
		}
		last = f.TimestampMS
		_, _, ok := rule.Angle(f)
		if ok {
			valid++
			confSum += frameConfidence(f)
		}
	}
	confidence := 0.0
	if valid > 0 {
		confidence = confSum / float64(valid)
	}
	if float64(valid)/float64(len(in.Frames)) < 0.55 {
		return Result{}, errors.New("pose visibility is too low")
	}
	reps := detectReps(in.Frames, rule)
	romScore := scoreROM(reps, rule.MinROM)
	tempoScore := scoreTempo(reps, rule.MinRepMS, rule.MaxRepMS)
	symmetryScore := scoreSymmetry(reps)
	stabilityScore := scoreStability(in.Frames, rule)
	techniqueScore := int(math.Round(0.35*float64(romScore) + 0.25*float64(tempoScore) + 0.25*float64(symmetryScore) + 0.15*float64(stabilityScore)))
	if techniqueScore < 0 {
		techniqueScore = 0
	}
	if techniqueScore > 100 {
		techniqueScore = 100
	}
	duration := in.DurationMS
	if duration <= 0 {
		duration = in.Frames[len(in.Frames)-1].TimestampMS - in.Frames[0].TimestampMS
	}
	avgEccentric, avgConcentric, avgRPM := phaseAverages(reps)
	result := Result{ExerciseKey: key, ExerciseName: rule.Name, CaptureMode: mode, WorkoutID: in.WorkoutID, WorkoutExerciseID: in.WorkoutExerciseID, SetNumber: in.SetNumber, RepCount: len(reps), TechniqueScore: techniqueScore, ROMScore: romScore, TempoScore: tempoScore, SymmetryScore: symmetryScore, StabilityScore: stabilityScore, Confidence: round(confidence, 3), DurationMS: duration, AverageEccentricMS: avgEccentric, AverageConcentricMS: avgConcentric, AverageRepRPM: avgRPM, AlgorithmVersion: AlgorithmVersion, Reps: reps, Feedback: feedback(reps, romScore, tempoScore, symmetryScore, stabilityScore, confidence), CreatedAt: time.Now().UTC()}
	raw, _ := json.Marshal(result)
	saved, err := s.st.SaveTechniqueAnalysis(ctx, store.TechniqueAnalysis{UserID: userID, ExerciseKey: key, CaptureMode: mode, WorkoutID: in.WorkoutID, WorkoutExerciseID: in.WorkoutExerciseID, SetNumber: in.SetNumber, RepCount: result.RepCount, TechniqueScore: techniqueScore, ROMScore: romScore, TempoScore: tempoScore, SymmetryScore: symmetryScore, StabilityScore: stabilityScore, Confidence: confidence, DurationMS: duration, AlgorithmVersion: AlgorithmVersion, ResultJSON: string(raw), CreatedAt: result.CreatedAt})
	if err != nil {
		return Result{}, err
	}
	result.ID = saved.ID
	raw, _ = json.Marshal(result)
	_ = s.st.UpdateTechniqueAnalysisResult(ctx, userID, result.ID, string(raw))
	return result, nil
}

func (s *Service) validateWorkoutLink(ctx context.Context, userID, techniqueKey string, in AnalyzeInput) error {
	linked := in.WorkoutID != "" || in.WorkoutExerciseID != "" || in.SetNumber != 0
	if !linked {
		return nil
	}
	if in.WorkoutID == "" || in.WorkoutExerciseID == "" || in.SetNumber < 1 {
		return errors.New("workout_id, workout_exercise_id and set_number must be provided together")
	}
	details, err := s.st.GetWorkout(ctx, userID, in.WorkoutID)
	if err != nil {
		return err
	}
	if details.Workout.Status != "active" {
		return errors.New("technique analysis can only be linked to an active workout")
	}
	for _, item := range details.Exercises {
		if item.ID != in.WorkoutExerciseID {
			continue
		}
		if in.SetNumber > item.TargetSets {
			return errors.New("set_number exceeds planned sets")
		}
		expected := techniqueKeyForExerciseID(item.ExerciseID)
		if expected == "" {
			return errors.New("workout exercise is not supported by technique analysis")
		}
		if expected != techniqueKey {
			return errors.New("technique exercise does not match workout exercise")
		}
		return nil
	}
	return errors.New("workout exercise not found")
}

func techniqueKeyForExerciseID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	switch {
	case id == "pushup" || strings.Contains(id, "push_up"):
		return "push_up"
	case strings.Contains(id, "squat") && !strings.Contains(id, "split"):
		return "squat"
	case strings.Contains(id, "split_squat") || strings.Contains(id, "lunge"):
		return "lunge"
	case strings.Contains(id, "curl"):
		return "biceps_curl"
	case strings.Contains(id, "overhead_press") || strings.Contains(id, "shoulder_press"):
		return "shoulder_press"
	default:
		return ""
	}
}

func (s *Service) Get(ctx context.Context, userID, id string) (Result, error) {
	a, err := s.st.GetTechniqueAnalysis(ctx, userID, id)
	if err != nil {
		return Result{}, err
	}
	var r Result
	if err = json.Unmarshal([]byte(a.ResultJSON), &r); err != nil {
		return Result{}, err
	}
	r.ID = a.ID
	normalizePersistedResult(&r, a)
	return r, nil
}
func (s *Service) List(ctx context.Context, userID string, limit int) ([]Result, error) {
	as, err := s.st.ListTechniqueAnalyses(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]Result, 0, len(as))
	for _, a := range as {
		var r Result
		if json.Unmarshal([]byte(a.ResultJSON), &r) == nil {
			r.ID = a.ID
			normalizePersistedResult(&r, a)
			out = append(out, r)
		}
	}
	return out, nil
}

func normalizePersistedResult(r *Result, a store.TechniqueAnalysis) {
	if r.CaptureMode == "" {
		r.CaptureMode = a.CaptureMode
	}
	if r.CaptureMode == "" {
		r.CaptureMode = "recorded"
	}
	if r.WorkoutID == "" {
		r.WorkoutID = a.WorkoutID
	}
	if r.WorkoutExerciseID == "" {
		r.WorkoutExerciseID = a.WorkoutExerciseID
	}
	if r.SetNumber == 0 {
		r.SetNumber = a.SetNumber
	}
	if r.AlgorithmVersion == "" {
		r.AlgorithmVersion = a.AlgorithmVersion
	}
}

func detectReps(frames []Frame, r exerciseRule) []RepMetric {
	ready, inRep, hitTurn := false, false, false
	startMS := int64(0)
	turnMS := int64(0)
	minAvg, maxAvg := 999.0, 0.0
	minL, minR, maxL, maxR := 999.0, 999.0, 0.0, 0.0
	out := []RepMetric{}
	reset := func() { minAvg, maxAvg, minL, minR, maxL, maxR = 999, 0, 999, 999, 0, 0 }
	for _, f := range frames {
		l, rr, ok := r.Angle(f)
		if !ok {
			continue
		}
		avg := (l + rr) / 2
		if !inRep {
			if r.StartAt == "bottom" {
				if avg <= r.Bottom {
					ready = true
				}
				if ready && avg > r.Bottom+10 {
					inRep, hitTurn, startMS, turnMS = true, false, f.TimestampMS, 0
					reset()
				} else {
					continue
				}
			} else {
				if avg >= r.Top {
					ready = true
				}
				if ready && avg < r.Top-10 {
					inRep, hitTurn, startMS, turnMS = true, false, f.TimestampMS, 0
					reset()
				} else {
					continue
				}
			}
		}
		if avg < minAvg {
			minAvg = avg
		}
		if avg > maxAvg {
			maxAvg = avg
		}
		if l < minL {
			minL = l
		}
		if l > maxL {
			maxL = l
		}
		if rr < minR {
			minR = rr
		}
		if rr > maxR {
			maxR = rr
		}

		complete := false
		if r.StartAt == "bottom" {
			if avg >= r.Top {
				if !hitTurn {
					turnMS = f.TimestampMS
				}
				hitTurn = true
			}
			complete = hitTurn && avg <= r.Bottom
		} else {
			if avg <= r.Bottom {
				if !hitTurn {
					turnMS = f.TimestampMS
				}
				hitTurn = true
			}
			complete = hitTurn && avg >= r.Top
		}
		if !complete {
			continue
		}

		endMS := f.TimestampMS
		duration := endMS - startMS
		if duration >= 350 && duration <= 10000 && turnMS > startMS && turnMS < endMS {
			leftROM, rightROM := maxL-minL, maxR-minR
			firstPhase := turnMS - startMS
			returnPhase := endMS - turnMS
			eccentric, concentric := firstPhase, returnPhase
			if r.FirstPhase == "concentric" {
				concentric, eccentric = firstPhase, returnPhase
			}
			out = append(out, RepMetric{
				Number: len(out) + 1, StartMS: startMS, EndMS: endMS, DurationMS: duration,
				EccentricMS: eccentric, ConcentricMS: concentric,
				ROMDegrees: round(maxAvg-minAvg, 1), LeftROMDegrees: round(leftROM, 1), RightROMDegrees: round(rightROM, 1), SymmetryDelta: round(math.Abs(leftROM-rightROM), 1),
			})
		}
		inRep, hitTurn, ready = false, false, true
		reset()
	}
	return out
}

func phaseAverages(reps []RepMetric) (int64, int64, float64) {
	if len(reps) == 0 {
		return 0, 0, 0
	}
	var eccentric, concentric, total int64
	for _, rep := range reps {
		eccentric += rep.EccentricMS
		concentric += rep.ConcentricMS
		total += rep.DurationMS
	}
	avgE := eccentric / int64(len(reps))
	avgC := concentric / int64(len(reps))
	avgTotal := float64(total) / float64(len(reps))
	rpm := 0.0
	if avgTotal > 0 {
		rpm = round(60000.0/avgTotal, 1)
	}
	return avgE, avgC, rpm
}

func scoreROM(reps []RepMetric, min float64) int {
	if len(reps) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range reps {
		sum += r.ROMDegrees
	}
	avg := sum / float64(len(reps))
	return clamp(int(math.Round(avg/min*90+10)), 0, 100)
}
func scoreTempo(reps []RepMetric, min, max int64) int {
	if len(reps) == 0 {
		return 0
	}
	good := 0
	for _, r := range reps {
		if r.DurationMS >= min && r.DurationMS <= max {
			good++
		}
	}
	return clamp(int(math.Round(float64(good)/float64(len(reps))*100)), 0, 100)
}
func scoreSymmetry(reps []RepMetric) int {
	if len(reps) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range reps {
		sum += r.SymmetryDelta
	}
	d := sum / float64(len(reps))
	return clamp(int(math.Round(100-d*3)), 0, 100)
}
func scoreStability(frames []Frame, r exerciseRule) int {
	if r.Stability == nil {
		return 85
	}
	vals := []float64{}
	for _, f := range frames {
		v := r.Stability(f)
		if v >= 0 {
			vals = append(vals, v)
		}
	}
	if len(vals) == 0 {
		return 50
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return clamp(int(math.Round(sum/float64(len(vals)))), 0, 100)
}
func feedback(reps []RepMetric, rom, tempo, sym, stable int, conf float64) []string {
	out := []string{}
	if len(reps) == 0 {
		return []string{"Не удалось уверенно выделить полный цикл повторения. Проверьте ракурс и выполните движение в полной амплитуде."}
	}
	if rom < 75 {
		out = append(out, "Амплитуда части повторений ограничена — попробуйте сохранить контролируемую полную амплитуду.")
	} else {
		out = append(out, "Амплитуда движения стабильная.")
	}
	if tempo < 75 {
		out = append(out, "Темп повторений заметно меняется. Старайтесь выполнять повторения более равномерно.")
	}
	if sym < 80 {
		out = append(out, "Есть заметная разница между левой и правой сторонами. Снимите ещё один подход под тем же ракурсом для подтверждения.")
	}
	if stable < 75 {
		out = append(out, "Положение корпуса меняется сильнее желаемого. Уменьшите скорость и удерживайте корпус стабильнее.")
	}
	if conf < 0.7 {
		out = append(out, "Уверенность распознавания средняя: улучшите освещение и поместите всё тело в кадр.")
	}
	if len(out) == 1 && rom >= 75 && tempo >= 75 && sym >= 80 && stable >= 75 {
		out = append(out, "Техника по измеряемым параметрам выглядит последовательной.")
	}
	return out
}

func frameConfidence(f Frame) float64 {
	idx := []int{11, 12, 13, 14, 15, 16, 23, 24, 25, 26, 27, 28}
	sum := 0.0
	n := 0
	for _, i := range idx {
		if i < len(f.Landmarks) {
			sum += f.Landmarks[i].Visibility
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}
func visible(p Landmark) bool { return p.Visibility >= 0.35 }
func angle(a, b, c Landmark) (float64, bool) {
	if !visible(a) || !visible(b) || !visible(c) {
		return 0, false
	}
	v1x, v1y := a.X-b.X, a.Y-b.Y
	v2x, v2y := c.X-b.X, c.Y-b.Y
	den := math.Hypot(v1x, v1y) * math.Hypot(v2x, v2y)
	if den < 1e-8 {
		return 0, false
	}
	cos := (v1x*v2x + v1y*v2y) / den
	if cos > 1 {
		cos = 1
	}
	if cos < -1 {
		cos = -1
	}
	return math.Acos(cos) * 180 / math.Pi, true
}
func pairAngles(f Frame, a1, b1, c1, a2, b2, c2 int) (float64, float64, bool) {
	if len(f.Landmarks) < 33 {
		return 0, 0, false
	}
	l, ok1 := angle(f.Landmarks[a1], f.Landmarks[b1], f.Landmarks[c1])
	r, ok2 := angle(f.Landmarks[a2], f.Landmarks[b2], f.Landmarks[c2])
	return l, r, ok1 && ok2
}
func kneeAngles(f Frame) (float64, float64, bool)  { return pairAngles(f, 23, 25, 27, 24, 26, 28) }
func elbowAngles(f Frame) (float64, float64, bool) { return pairAngles(f, 11, 13, 15, 12, 14, 16) }
func lungeAngles(f Frame) (float64, float64, bool) {
	l, r, ok := kneeAngles(f)
	if !ok {
		return 0, 0, false
	}
	v := math.Min(l, r)
	return v, v, true
}
func trunkStability(f Frame) float64 {
	if len(f.Landmarks) < 29 {
		return -1
	}
	l, ok1 := angle(f.Landmarks[11], f.Landmarks[23], f.Landmarks[27])
	r, ok2 := angle(f.Landmarks[12], f.Landmarks[24], f.Landmarks[28])
	if !ok1 || !ok2 {
		return -1
	}
	dev := (math.Abs(180-l) + math.Abs(180-r)) / 2
	return math.Max(0, 100-dev*2)
}
func plankStability(f Frame) float64 { return trunkStability(f) }
func shoulderStability(f Frame) float64 {
	if len(f.Landmarks) < 25 {
		return -1
	}
	dy := math.Abs(f.Landmarks[11].Y - f.Landmarks[12].Y)
	return math.Max(0, 100-dy*500)
}
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func round(v float64, n int) float64 { p := math.Pow10(n); return math.Round(v*p) / p }
