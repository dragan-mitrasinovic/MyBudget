package httpapi

import (
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

// Config contains the HTTP-layer settings supplied by the executable.
type Config struct {
	GoogleClientID string
	SessionTTL     time.Duration
	// IsDevelopment enables dev-only routes and allows session cookies over localhost HTTP.
	IsDevelopment bool
}

// Server holds the dependencies and config shared by the HTTP handlers.
type Server struct {
	store            *store.Store
	googleClientID   string
	useSecureCookies bool
	sessionTTL       time.Duration
}
