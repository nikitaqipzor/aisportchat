//go:build cgo

package store

import (
	"context"
	"strconv"
	"strings"
	"time"
)

func (p *Postgres) CreateBodyScan(ctx context.Context, scan BodyScan) (BodyScanDetails, error) {
	if scan.ID == "" {
		scan.ID = newID()
	}
	if scan.Status == "" {
		scan.Status = "draft"
	}
	if scan.CreatedAt.IsZero() {
		scan.CreatedAt = time.Now().UTC()
	}
	rows, err := p.query(ctx, `INSERT INTO body_scans(id,user_id,status,created_at,completed_at) VALUES($1::uuid,$2::uuid,$3,$4::timestamptz,$5::timestamptz) RETURNING id::text,user_id::text,status,created_at::text,completed_at::text`, sp(scan.ID), sp(scan.UserID), sp(scan.Status), sp(scan.CreatedAt.Format(time.RFC3339Nano)), tp(scan.CompletedAt))
	if err != nil {
		return BodyScanDetails{}, err
	}
	parsed, err := scanBodyScan(rows[0])
	if err != nil {
		return BodyScanDetails{}, err
	}
	return BodyScanDetails{Scan: parsed, Photos: []BodyScanPhoto{}}, nil
}

func (p *Postgres) GetBodyScan(ctx context.Context, userID, scanID string) (BodyScanDetails, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,status,created_at::text,completed_at::text FROM body_scans WHERE id=$1::uuid AND user_id=$2::uuid LIMIT 1`, sp(scanID), sp(userID))
	if err != nil {
		return BodyScanDetails{}, err
	}
	if len(rows) == 0 {
		return BodyScanDetails{}, ErrNotFound
	}
	scan, err := scanBodyScan(rows[0])
	if err != nil {
		return BodyScanDetails{}, err
	}
	pr, err := p.query(ctx, `SELECT p.id::text,p.scan_id::text,s.user_id::text,p.view,p.storage_key,p.mime_type,p.width::text,p.height::text,p.bytes::text,p.brightness::text,p.contrast::text,p.quality_status,p.quality_issues,p.created_at::text FROM body_scan_photos p JOIN body_scans s ON s.id=p.scan_id WHERE p.scan_id=$1::uuid ORDER BY CASE p.view WHEN 'front' THEN 1 WHEN 'side' THEN 2 ELSE 3 END`, sp(scanID))
	if err != nil {
		return BodyScanDetails{}, err
	}
	photos := make([]BodyScanPhoto, 0, len(pr))
	for _, r := range pr {
		ph, e := scanBodyScanPhoto(r)
		if e != nil {
			return BodyScanDetails{}, e
		}
		photos = append(photos, ph)
	}
	return BodyScanDetails{Scan: scan, Photos: photos}, nil
}

func (p *Postgres) ListBodyScans(ctx context.Context, userID string, limit int) ([]BodyScanDetails, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := p.query(ctx, `SELECT id::text FROM body_scans WHERE user_id=$1::uuid ORDER BY created_at DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]BodyScanDetails, 0, len(rows))
	for _, r := range rows {
		d, e := p.GetBodyScan(ctx, userID, val(r, 0))
		if e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, nil
}

func (p *Postgres) UpsertBodyScanPhoto(ctx context.Context, photo BodyScanPhoto) (BodyScanDetails, error) {
	if photo.ID == "" {
		photo.ID = newID()
	}
	if photo.CreatedAt.IsZero() {
		photo.CreatedAt = time.Now().UTC()
	}
	issues := strings.Join(photo.QualityIssues, "|")
	rows, err := p.query(ctx, `INSERT INTO body_scan_photos(id,scan_id,view,storage_key,mime_type,width,height,bytes,brightness,contrast,quality_status,quality_issues,created_at)
VALUES($1::uuid,$2::uuid,$3,$4,$5,$6::int,$7::int,$8::int,$9::numeric,$10::numeric,$11,$12,$13::timestamptz)
ON CONFLICT(scan_id,view) DO UPDATE SET id=EXCLUDED.id,storage_key=EXCLUDED.storage_key,mime_type=EXCLUDED.mime_type,width=EXCLUDED.width,height=EXCLUDED.height,bytes=EXCLUDED.bytes,brightness=EXCLUDED.brightness,contrast=EXCLUDED.contrast,quality_status=EXCLUDED.quality_status,quality_issues=EXCLUDED.quality_issues,created_at=EXCLUDED.created_at
RETURNING id::text`, sp(photo.ID), sp(photo.ScanID), sp(photo.View), sp(photo.StorageKey), sp(photo.MimeType), sp(strconv.Itoa(photo.Width)), sp(strconv.Itoa(photo.Height)), sp(strconv.Itoa(photo.Bytes)), sp(strconv.FormatFloat(photo.Brightness, 'f', 2, 64)), sp(strconv.FormatFloat(photo.Contrast, 'f', 2, 64)), sp(photo.QualityStatus), sp(issues), sp(photo.CreatedAt.Format(time.RFC3339Nano)))
	if err != nil {
		return BodyScanDetails{}, err
	}
	if len(rows) == 0 {
		return BodyScanDetails{}, ErrNotFound
	}
	return p.GetBodyScan(ctx, photo.UserID, photo.ScanID)
}

func (p *Postgres) CompleteBodyScan(ctx context.Context, userID, scanID string, completedAt time.Time) (BodyScanDetails, error) {
	rows, err := p.query(ctx, `UPDATE body_scans s SET status='completed',completed_at=$1::timestamptz WHERE s.id=$2::uuid AND s.user_id=$3::uuid AND s.status='draft' AND (SELECT count(*) FROM body_scan_photos p WHERE p.scan_id=s.id)=3 RETURNING s.id::text`, sp(completedAt.UTC().Format(time.RFC3339Nano)), sp(scanID), sp(userID))
	if err != nil {
		return BodyScanDetails{}, err
	}
	if len(rows) == 0 {
		d, e := p.GetBodyScan(ctx, userID, scanID)
		if e != nil {
			return BodyScanDetails{}, e
		}
		if d.Scan.Status == "completed" {
			return d, nil
		}
		return BodyScanDetails{}, ErrInvalidState
	}
	return p.GetBodyScan(ctx, userID, scanID)
}

func scanBodyScan(r []*string) (BodyScan, error) {
	created, err := parseTime(val(r, 3))
	if err != nil {
		return BodyScan{}, err
	}
	return BodyScan{ID: val(r, 0), UserID: val(r, 1), Status: val(r, 2), CreatedAt: created, CompletedAt: timePtr(r[4])}, nil
}
func scanBodyScanPhoto(r []*string) (BodyScanPhoto, error) {
	w, _ := strconv.Atoi(val(r, 6))
	h, _ := strconv.Atoi(val(r, 7))
	size, _ := strconv.Atoi(val(r, 8))
	bright, _ := strconv.ParseFloat(val(r, 9), 64)
	contrast, _ := strconv.ParseFloat(val(r, 10), 64)
	created, err := parseTime(val(r, 13))
	if err != nil {
		return BodyScanPhoto{}, err
	}
	issues := []string{}
	if raw := val(r, 12); raw != "" {
		issues = strings.Split(raw, "|")
	}
	return BodyScanPhoto{ID: val(r, 0), ScanID: val(r, 1), UserID: val(r, 2), View: val(r, 3), StorageKey: val(r, 4), MimeType: val(r, 5), Width: w, Height: h, Bytes: size, Brightness: bright, Contrast: contrast, QualityStatus: val(r, 11), QualityIssues: issues, CreatedAt: created}, nil
}

func (p *Postgres) DeleteBodyScan(ctx context.Context, userID, scanID string) error {
	rows, err := p.query(ctx, `DELETE FROM body_scans WHERE id=$1::uuid AND user_id=$2::uuid RETURNING id::text`, sp(scanID), sp(userID))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return ErrNotFound
	}
	return nil
}
