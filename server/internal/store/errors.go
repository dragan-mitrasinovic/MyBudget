package store

import (
	"errors"
	"strings"
)

// Sentinel errors mapped to HTTP status codes by the transport layer.
var (
	ErrNotFound               = errors.New("not found")
	ErrDuplicateName          = errors.New("name already exists")
	ErrInvalidCategory        = errors.New("category does not belong to this budget")
	ErrCategoryNotEmpty       = errors.New("category has expenses")
	ErrInvalidReassignment    = errors.New("invalid reassignment category")
	ErrInvalidDashboardLayout = errors.New("invalid dashboard layout")
)

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
