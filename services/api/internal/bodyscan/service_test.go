package bodyscan

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type deleteFailingMedia struct {
	media.Store
	failKey string
}

func (m *deleteFailingMedia) Delete(ctx context.Context, key string) error {
	if key == m.failKey {
		return errors.New("injected blob delete failure")
	}
	return m.Store.Delete(ctx, key)
}

func testImage(t *testing.T, dark bool) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 900, 1400))
	for y := 0; y < 1400; y++ {
		for x := 0; x < 900; x++ {
			v := uint8(0)
			if !dark {
				v = uint8(55 + ((x + y) % 150))
			}
			img.Set(x, y, color.RGBA{v, uint8(minInt(255, int(v)+12)), uint8(maxInt(0, int(v)-8)), 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatal(err)
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func testUser(t *testing.T, st *store.Memory, email string) string {
	t.Helper()
	if _, err := st.CreateUser(context.Background(), email, "hash"); err != nil {
		t.Fatal(err)
	}
	u, err := st.FindUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func TestBodyScanFlowAndPrivacy(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	blobs := media.NewMemoryStore()
	svc := NewService(st, blobs)
	user := testUser(t, st, "one@example.com")
	other := testUser(t, st, "other@example.com")
	scan, err := svc.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Complete(ctx, user, scan.Scan.ID); err == nil {
		t.Fatal("expected incomplete scan to fail")
	}
	for _, view := range []string{"front", "side", "back"} {
		scan, err = svc.AddPhoto(ctx, user, scan.Scan.ID, view, testImage(t, false))
		if err != nil {
			t.Fatalf("%s: %v", view, err)
		}
	}
	if len(scan.Photos) != 3 {
		t.Fatalf("got %d photos", len(scan.Photos))
	}
	done, err := svc.Complete(ctx, user, scan.Scan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Scan.Status != "completed" {
		t.Fatalf("status=%s", done.Scan.Status)
	}
	if _, err := svc.Get(ctx, other, scan.Scan.ID); err != store.ErrNotFound {
		t.Fatalf("privacy: %v", err)
	}
	blob, err := svc.Photo(ctx, user, scan.Scan.ID, "front")
	if err != nil || len(blob.Bytes) == 0 {
		t.Fatalf("photo: %v", err)
	}
	frontKey := done.Photos[0].StorageKey
	if err := svc.Delete(ctx, user, scan.Scan.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, user, scan.Scan.ID); err != store.ErrNotFound {
		t.Fatalf("deleted metadata still accessible: %v", err)
	}
	if _, err := blobs.Get(ctx, frontKey); err != media.ErrNotFound {
		t.Fatalf("deleted media still accessible: %v", err)
	}
}

func TestDeleteKeepsMetadataWhenPrivateBlobDeletionFails(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	blobs := media.NewMemoryStore()
	user := testUser(t, st, "delete-retry@example.com")
	svc := NewService(st, blobs)
	scan, err := svc.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range []string{"front", "side", "back"} {
		scan, err = svc.AddPhoto(ctx, user, scan.Scan.ID, view, testImage(t, false))
		if err != nil {
			t.Fatal(err)
		}
	}
	failing := &deleteFailingMedia{Store: blobs, failKey: scan.Photos[1].StorageKey}
	if err := NewService(st, failing).Delete(ctx, user, scan.Scan.ID); err == nil {
		t.Fatal("expected private blob deletion failure")
	}
	if _, err := svc.Get(ctx, user, scan.Scan.ID); err != nil {
		t.Fatalf("metadata must remain available for retry: %v", err)
	}
	failing.failKey = ""
	if err := NewService(st, failing).Delete(ctx, user, scan.Scan.ID); err != nil {
		t.Fatalf("idempotent retry: %v", err)
	}
	if _, err := svc.Get(ctx, user, scan.Scan.ID); err != store.ErrNotFound {
		t.Fatalf("metadata still accessible after retry: %v", err)
	}
}

func TestBodyScanRejectsDarkImage(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	user := testUser(t, st, "dark@example.com")
	svc := NewService(st, media.NewMemoryStore())
	scan, _ := svc.Create(ctx, user)
	_, err := svc.AddPhoto(ctx, user, scan.Scan.ID, "front", testImage(t, true))
	if err == nil || !strings.Contains(err.Error(), "too dark") {
		t.Fatalf("expected dark rejection, got %v", err)
	}
}

func TestLatestComparisonUsesCaptureAndMeasurements(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	svc := NewService(st, media.NewMemoryStore())
	user := testUser(t, st, "compare@example.com")
	now := time.Now().UTC()
	create := func(id string, at time.Time) {
		d, err := st.CreateBodyScan(ctx, store.BodyScan{ID: id, UserID: user, Status: "draft", CreatedAt: at})
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range []string{"front", "side", "back"} {
			d, err = svc.AddPhoto(ctx, user, d.Scan.ID, v, testImage(t, false))
			if err != nil {
				t.Fatal(err)
			}
		}
		if _, err = st.CompleteBodyScan(ctx, user, d.Scan.ID, at.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	create("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", now.AddDate(0, 0, -30))
	create("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", now)
	w1, w2 := 90.0, 88.5
	wa1, wa2 := 92.0, 89.0
	_, _ = st.CreateBodyMeasurement(ctx, store.BodyMeasurement{UserID: user, LoggedAt: now.AddDate(0, 0, -30), WeightKG: &w1, WaistCM: &wa1})
	_, _ = st.CreateBodyMeasurement(ctx, store.BodyMeasurement{UserID: user, LoggedAt: now, WeightKG: &w2, WaistCM: &wa2})
	cmp, err := svc.LatestComparison(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if cmp.DaysBetween < 29 || cmp.CaptureScore < 90 {
		t.Fatalf("cmp=%+v", cmp)
	}
	if cmp.CaptureGrade != "excellent" || cmp.VisualCVStatus != "capture_only_no_body_inference" {
		t.Fatalf("unexpected deterministic classification: %+v", cmp)
	}
	if len(cmp.ViewMetrics) != 3 {
		t.Fatalf("view metrics=%+v", cmp.ViewMetrics)
	}
	for _, metric := range cmp.ViewMetrics {
		if metric.ConsistencyScore != 100 || metric.BrightnessDelta != 0 || metric.ContrastDelta != 0 || metric.ResolutionDeltaPercent != 0 {
			t.Fatalf("identical capture metric=%+v", metric)
		}
	}
	if cmp.WeightDeltaKG == nil || *cmp.WeightDeltaKG != -1.5 {
		t.Fatalf("weight delta=%v", cmp.WeightDeltaKG)
	}
}

func TestCompareViewsReportsTechnicalDifferencesOnly(t *testing.T) {
	from := []store.BodyScanPhoto{
		{View: "front", Width: 1000, Height: 1500, Brightness: 100, Contrast: 40},
		{View: "side", Width: 1000, Height: 1500, Brightness: 100, Contrast: 40},
		{View: "back", Width: 1000, Height: 1500, Brightness: 100, Contrast: 40},
	}
	to := []store.BodyScanPhoto{
		{View: "front", Width: 800, Height: 1200, Brightness: 130, Contrast: 60},
		{View: "side", Width: 1000, Height: 1500, Brightness: 100, Contrast: 40},
		{View: "back", Width: 1000, Height: 1500, Brightness: 100, Contrast: 40},
	}
	metrics := compareViews(from, to)
	if len(metrics) != 3 {
		t.Fatalf("metrics=%+v", metrics)
	}
	if metrics[0].BrightnessDelta != 30 || metrics[0].ContrastDelta != 20 || metrics[0].ResolutionDeltaPercent != 20 {
		t.Fatalf("front=%+v", metrics[0])
	}
	if metrics[0].ConsistencyScore >= metrics[1].ConsistencyScore {
		t.Fatalf("changed capture must score below unchanged: %+v", metrics)
	}
}
