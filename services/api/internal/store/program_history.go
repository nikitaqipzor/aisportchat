package store

import "time"

// ProgramHistoryCursor identifies the final program on a history page.
type ProgramHistoryCursor struct {
	UpdatedAt time.Time
	ID        string
}
