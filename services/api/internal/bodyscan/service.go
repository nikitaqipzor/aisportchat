package bodyscan

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

var (
	ErrInvalidImage = errors.New("invalid image")
	ErrQuality      = errors.New("image quality check failed")
)

const maxImageBytes = 8 * 1024 * 1024

type Service struct {
	store store.Store
	media media.Store
}

func NewService(st store.Store, blobs media.Store) *Service { return &Service{store: st, media: blobs} }

func (s *Service) Create(ctx context.Context, userID string) (store.BodyScanDetails, error) {
	items, err := s.store.ListBodyScans(ctx, userID, 20)
	if err == nil {
		for _, item := range items {
			if item.Scan.Status == "draft" {
				return item, nil
			}
		}
	}
	return s.store.CreateBodyScan(ctx, store.BodyScan{ID: newID(), UserID: userID, Status: "draft", CreatedAt: time.Now().UTC()})
}
func (s *Service) Get(ctx context.Context, userID, scanID string) (store.BodyScanDetails, error) {
	return s.store.GetBodyScan(ctx, userID, scanID)
}
func (s *Service) List(ctx context.Context, userID string, limit int) ([]store.BodyScanDetails, error) {
	return s.store.ListBodyScans(ctx, userID, limit)
}

func (s *Service) AddPhoto(ctx context.Context, userID, scanID, view, dataURL string) (store.BodyScanDetails, error) {
	if view != "front" && view != "side" && view != "back" {
		return store.BodyScanDetails{}, fmt.Errorf("%w: view must be front, side or back", ErrInvalidImage)
	}
	details, err := s.store.GetBodyScan(ctx, userID, scanID)
	if err != nil {
		return store.BodyScanDetails{}, err
	}
	if details.Scan.Status == "completed" {
		return store.BodyScanDetails{}, store.ErrInvalidState
	}
	data, mime, err := decodeDataURL(dataURL)
	if err != nil {
		return store.BodyScanDetails{}, err
	}
	quality, err := analyze(data)
	if err != nil {
		return store.BodyScanDetails{}, err
	}
	if quality.Status == "rejected" {
		return store.BodyScanDetails{}, fmt.Errorf("%w: %s", ErrQuality, strings.Join(quality.Issues, ", "))
	}
	ext := "jpg"
	if mime == "image/png" {
		ext = "png"
	}
	key := fmt.Sprintf("body-scans/%s/%s/%s.%s", userID, scanID, view, ext)
	if err := s.media.Put(ctx, key, mime, data); err != nil {
		return store.BodyScanDetails{}, err
	}
	photo := store.BodyScanPhoto{ID: newID(), ScanID: scanID, UserID: userID, View: view, StorageKey: key, MimeType: mime, Width: quality.Width, Height: quality.Height, Bytes: len(data), Brightness: quality.Brightness, Contrast: quality.Contrast, QualityStatus: quality.Status, QualityIssues: quality.Issues, CreatedAt: time.Now().UTC()}
	out, err := s.store.UpsertBodyScanPhoto(ctx, photo)
	if err != nil {
		_ = s.media.Delete(ctx, key)
		return store.BodyScanDetails{}, err
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, userID, scanID string) error {
	details, err := s.store.GetBodyScan(ctx, userID, scanID)
	if err != nil {
		return err
	}
	for _, photo := range details.Photos {
		_ = s.media.Delete(ctx, photo.StorageKey)
	}
	return s.store.DeleteBodyScan(ctx, userID, scanID)
}

func (s *Service) Complete(ctx context.Context, userID, scanID string) (store.BodyScanDetails, error) {
	return s.store.CompleteBodyScan(ctx, userID, scanID, time.Now().UTC())
}

func (s *Service) Photo(ctx context.Context, userID, scanID, view string) (media.Blob, error) {
	d, err := s.store.GetBodyScan(ctx, userID, scanID)
	if err != nil {
		return media.Blob{}, err
	}
	for _, p := range d.Photos {
		if p.View == view {
			return s.media.Get(ctx, p.StorageKey)
		}
	}
	return media.Blob{}, store.ErrNotFound
}

type Quality struct {
	Status               string
	Issues               []string
	Width, Height        int
	Brightness, Contrast float64
}

func analyze(data []byte) (Quality, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Quality{}, fmt.Errorf("%w: image decode failed", ErrInvalidImage)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	q := Quality{Status: "accepted", Width: w, Height: h}
	hard := []string{}
	warn := []string{}
	if w < 720 || h < 960 {
		hard = append(hard, "resolution too low")
	}
	if h <= w {
		hard = append(hard, "use portrait orientation")
	}
	ratio := float64(w) / float64(h)
	if ratio < 0.48 || ratio > 0.85 {
		warn = append(warn, "unusual framing; keep full body centered")
	}
	step := int(math.Sqrt(float64(w*h) / 12000))
	if step < 1 {
		step = 1
	}
	var sum, sumSq float64
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			r, g, bl, _ := img.At(x, y).RGBA()
			lum := 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(bl>>8)
			sum += lum
			sumSq += lum * lum
			n++
		}
	}
	if n > 0 {
		q.Brightness = sum / float64(n)
		variance := sumSq/float64(n) - q.Brightness*q.Brightness
		if variance > 0 {
			q.Contrast = math.Sqrt(variance)
		}
	}
	if q.Brightness < 35 {
		hard = append(hard, "image is too dark")
	}
	if q.Brightness > 235 {
		hard = append(hard, "image is overexposed")
	}
	if q.Contrast < 18 {
		warn = append(warn, "low contrast")
	}
	q.Brightness = math.Round(q.Brightness*10) / 10
	q.Contrast = math.Round(q.Contrast*10) / 10
	if len(hard) > 0 {
		q.Status = "rejected"
		q.Issues = append(hard, warn...)
	} else if len(warn) > 0 {
		q.Status = "warning"
		q.Issues = warn
	}
	return q, nil
}

func decodeDataURL(v string) ([]byte, string, error) {
	if len(v) > maxImageBytes*2 {
		return nil, "", fmt.Errorf("%w: image too large", ErrInvalidImage)
	}
	header, payload, ok := strings.Cut(v, ",")
	if !ok || !strings.HasPrefix(header, "data:image/") || !strings.Contains(header, ";base64") {
		return nil, "", fmt.Errorf("%w: expected base64 image data URL", ErrInvalidImage)
	}
	mime := strings.TrimPrefix(strings.Split(header, ";")[0], "data:")
	if mime != "image/jpeg" && mime != "image/png" {
		return nil, "", fmt.Errorf("%w: unsupported image type", ErrInvalidImage)
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid base64", ErrInvalidImage)
	}
	if len(data) == 0 || len(data) > maxImageBytes {
		return nil, "", fmt.Errorf("%w: image must be <= 8 MB", ErrInvalidImage)
	}
	return data, mime, nil
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type Comparison struct {
	FromScanID     string   `json:"from_scan_id"`
	ToScanID       string   `json:"to_scan_id"`
	DaysBetween    int      `json:"days_between"`
	CaptureScore   float64  `json:"capture_consistency_score"`
	LightingDelta  float64  `json:"lighting_delta"`
	Warnings       []string `json:"warnings,omitempty"`
	WeightDeltaKG  *float64 `json:"weight_delta_kg,omitempty"`
	WaistDeltaCM   *float64 `json:"waist_delta_cm,omitempty"`
	VisualCVStatus string   `json:"visual_cv_status"`
}

func (s *Service) LatestComparison(ctx context.Context, userID string) (Comparison, error) {
	scans, err := s.store.ListBodyScans(ctx, userID, 20)
	if err != nil {
		return Comparison{}, err
	}
	completed := make([]store.BodyScanDetails, 0, 2)
	for _, item := range scans {
		if item.Scan.Status == "completed" {
			completed = append(completed, item)
			if len(completed) == 2 {
				break
			}
		}
	}
	if len(completed) < 2 {
		return Comparison{}, store.ErrNotFound
	}
	to, from := completed[0], completed[1]
	toAt := to.Scan.CreatedAt
	if to.Scan.CompletedAt != nil {
		toAt = *to.Scan.CompletedAt
	}
	fromAt := from.Scan.CreatedAt
	if from.Scan.CompletedAt != nil {
		fromAt = *from.Scan.CompletedAt
	}
	lightingDelta := math.Abs(avgBrightness(to.Photos) - avgBrightness(from.Photos))
	dimensionPenalty := avgDimensionDelta(from.Photos, to.Photos)
	score := 100 - math.Min(45, lightingDelta*0.7) - math.Min(35, dimensionPenalty*100)
	warnings := []string{}
	if lightingDelta > 25 {
		warnings = append(warnings, "Освещение заметно отличается — визуальное сравнение будет менее точным.")
	}
	if dimensionPenalty > 0.18 {
		warnings = append(warnings, "Кадрирование отличается — старайтесь держать одинаковое расстояние до камеры.")
	}
	for _, p := range append(append([]store.BodyScanPhoto{}, from.Photos...), to.Photos...) {
		if p.QualityStatus == "warning" {
			warnings = append(warnings, "Один из кадров имеет предупреждение качества.")
			break
		}
	}
	if score < 0 {
		score = 0
	}
	score = math.Round(score*10) / 10
	cmp := Comparison{FromScanID: from.Scan.ID, ToScanID: to.Scan.ID, DaysBetween: int(toAt.Sub(fromAt).Hours() / 24), CaptureScore: score, LightingDelta: math.Round(lightingDelta*10) / 10, Warnings: warnings, VisualCVStatus: "pending_pose_engine"}
	measurements, _ := s.store.ListBodyMeasurements(ctx, userID, fromAt.Add(-72*time.Hour), toAt.Add(72*time.Hour))
	if a, b := nearestMeasurement(measurements, fromAt), nearestMeasurement(measurements, toAt); a != nil && b != nil {
		if a.WeightKG != nil && b.WeightKG != nil {
			v := math.Round((*b.WeightKG-*a.WeightKG)*100) / 100
			cmp.WeightDeltaKG = &v
		}
		if a.WaistCM != nil && b.WaistCM != nil {
			v := math.Round((*b.WaistCM-*a.WaistCM)*100) / 100
			cmp.WaistDeltaCM = &v
		}
	}
	return cmp, nil
}

func avgBrightness(photos []store.BodyScanPhoto) float64 {
	if len(photos) == 0 {
		return 0
	}
	var x float64
	for _, p := range photos {
		x += p.Brightness
	}
	return x / float64(len(photos))
}
func avgDimensionDelta(a, b []store.BodyScanPhoto) float64 {
	ma := map[string]store.BodyScanPhoto{}
	for _, p := range a {
		ma[p.View] = p
	}
	var total float64
	n := 0
	for _, p := range b {
		q, ok := ma[p.View]
		if !ok || q.Width == 0 || q.Height == 0 {
			continue
		}
		dw := math.Abs(float64(p.Width-q.Width)) / float64(q.Width)
		dh := math.Abs(float64(p.Height-q.Height)) / float64(q.Height)
		total += (dw + dh) / 2
		n++
	}
	if n == 0 {
		return 1
	}
	return total / float64(n)
}
func nearestMeasurement(items []store.BodyMeasurement, target time.Time) *store.BodyMeasurement {
	var best *store.BodyMeasurement
	bestDist := time.Duration(1<<63 - 1)
	for i := range items {
		d := items[i].LoggedAt.Sub(target)
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			best = &items[i]
		}
	}
	if bestDist > 96*time.Hour {
		return nil
	}
	return best
}
