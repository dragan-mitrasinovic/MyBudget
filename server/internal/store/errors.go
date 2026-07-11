package store

import (
	"errors"
	"strings"
)

// Sentinel errors mapped to HTTP status codes by the transport layer.
var (
	ErrNotFound        = errors.New("not found")
	ErrSystemCategory  = errors.New("cannot modify the system category")
	ErrDuplicateName   = errors.New("name already exists")
	ErrInvalidCategory = errors.New("category does not belong to this budget")
)

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
