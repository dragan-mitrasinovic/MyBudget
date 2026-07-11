package httpapi

import (
	"log"
	"net/http"
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

// New wires the API routes and returns the complete HTTP handler.
func New(st *store.Store, cfg Config) http.Handler {
	srv := newServer(st, cfg)
	requireAppOrigin := func(handler http.HandlerFunc) http.Handler {
		return srv.requireAppOrigin(handler)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /api/auth/google", requireAppOrigin(srv.handleGoogleLogin))
	mux.HandleFunc("GET /api/me", srv.handleGetCurrentUser)
	mux.Handle("POST /api/auth/logout", requireAppOrigin(srv.handleLogout))

	mux.HandleFunc("GET /api/budgets", srv.handleListBudgets)
	mux.Handle("POST /api/budgets", requireAppOrigin(srv.handleCreateBudget))
	mux.HandleFunc("GET /api/budgets/{id}", srv.handleGetBudget)
	mux.Handle("PATCH /api/budgets/{id}", requireAppOrigin(srv.handleUpdateBudget))
	mux.Handle("DELETE /api/budgets/{id}", requireAppOrigin(srv.handleDeleteBudget))

	mux.HandleFunc("GET /api/budgets/{id}/categories", srv.handleListCategories)
	mux.Handle("POST /api/budgets/{id}/categories", requireAppOrigin(srv.handleCreateCategory))
	mux.Handle("PATCH /api/categories/{id}", requireAppOrigin(srv.handleUpdateCategory))
	mux.Handle("DELETE /api/categories/{id}", requireAppOrigin(srv.handleDeleteCategory))

	mux.HandleFunc("GET /api/budgets/{id}/expenses", srv.handleListExpenses)
	mux.Handle("POST /api/budgets/{id}/expenses", requireAppOrigin(srv.handleCreateExpense))
	mux.Handle("PATCH /api/expenses/{id}", requireAppOrigin(srv.handleUpdateExpense))
	mux.Handle("DELETE /api/expenses/{id}", requireAppOrigin(srv.handleDeleteExpense))

	// The local and automated-testing shortcut is registered only when explicitly enabled.
	if cfg.IsDevelopment {
		log.Println("⚠  DEV AUTH ENABLED — /api/dev/login mints a session for a dummy user")
		mux.Handle("POST /api/dev/login", requireAppOrigin(srv.handleDevLogin))
		mux.HandleFunc("GET /api/dev/login", srv.handleDevLogin)
	}

	return mux
}

func newServer(st *store.Store, cfg Config) *Server {
	sessionRenewalInterval := 24 * time.Hour
	if cfg.SessionTTL <= sessionRenewalInterval {
		sessionRenewalInterval = cfg.SessionTTL / 2
	}

	return &Server{
		store:                  st,
		googleClientID:         cfg.GoogleClientID,
		useSecureCookies:       !cfg.IsDevelopment,
		sessionTTL:             cfg.SessionTTL,
		sessionRenewalInterval: sessionRenewalInterval,
		appOrigin:              cfg.AppOrigin,
	}
}
