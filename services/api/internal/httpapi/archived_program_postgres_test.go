//go:build cgo

package httpapi

import (
	"os"
	"testing"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestConcurrentProgramSessionWorkoutAndArchivePostgres(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is required for PostgreSQL concurrency integration")
	}
	pg, err := store.NewPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pg.Close)
	archiveDB, err := store.NewPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(archiveDB.Close)
	testConcurrentProgramSessionWorkoutAndArchive(t, pg, archiveDB)
}
