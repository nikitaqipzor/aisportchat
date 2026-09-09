//go:build !cgo

package store

import "errors"

// OpenPostgresStore keeps non-cgo builds compilable while failing clearly at
// runtime if PostgreSQL is selected without the libpq-backed adapter.
func OpenPostgresStore(string) (Store, func(), error) {
	return nil, func() {}, errors.New("postgres backend requires a cgo-enabled build with libpq")
}
