//go:build cgo

package store

import "context"

// OpenPostgresStore creates and validates the production PostgreSQL store.
// Keeping this behind a build-specific factory lets the rest of the service
// compile in environments where cgo/libpq are intentionally unavailable.
func OpenPostgresStore(dsn string) (Store, func(), error) {
	pg, err := NewPostgres(dsn)
	if err != nil {
		return nil, func() {}, err
	}
	if err := pg.Ping(context.Background()); err != nil {
		pg.Close()
		return nil, func() {}, err
	}
	return pg, pg.Close, nil
}
