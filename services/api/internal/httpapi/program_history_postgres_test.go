//go:build cgo

package httpapi

import (
	"os"
	"testing"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestProgramHistoryPagesBeyondFiftyPostgres(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" { t.Skip("POSTGRES_TEST_DSN is required") }
	pg, err := store.NewPostgres(dsn)
	if err != nil { t.Fatal(err) }
	t.Cleanup(pg.Close)
	testProgramHistoryPagesBeyondFifty(t, pg)
}
