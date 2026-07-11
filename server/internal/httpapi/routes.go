package httpapi

import (
	"log"
	"net/http"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

// New wires the API routes and returns the complete HTTP handler.
func New(st *store.Store, cfg Config) http.Handler {
	srv := newServer(st, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/google", srv.handleGoogleLogin)
	mux.HandleFunc("GET /api/me", srv.handleGetCurrentUser)
	mux.HandleFunc("POST /api/auth/logout", srv.handleLogout)

	mux.HandleFunc("GET /api/budgets", srv.handleListBudgets)
	mux.HandleFunc("POST /api/budgets", srv.handleCreateBudget)
	mux.HandleFunc("GET /api/budgets/{id}", srv.handleGetBudget)
	mux.HandleFunc("PATCH /api/budgets/{id}", srv.handleUpdateBudget)
	mux.HandleFunc("DELETE /api/budgets/{id}", srv.handleDeleteBudget)

	mux.HandleFunc("GET /api/budgets/{id}/categories", srv.handleListCategories)
	mux.HandleFunc("POST /api/budgets/{id}/categories", srv.handleCreateCategory)
	mux.HandleFunc("PATCH /api/categories/{id}", srv.handleUpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", srv.handleDeleteCategory)

	mux.HandleFunc("GET /api/budgets/{id}/expenses", srv.handleListExpenses)
	mux.HandleFunc("POST /api/budgets/{id}/expenses", srv.handleCreateExpense)
	mux.HandleFunc("PATCH /api/expenses/{id}", srv.handleUpdateExpense)
	mux.HandleFunc("DELETE /api/expenses/{id}", srv.handleDeleteExpense)

	// The local and automated-testing shortcut is registered only when explicitly enabled.
	if cfg.IsDevelopment {
		log.Println("⚠  DEV AUTH ENABLED — /api/dev/login mints a session for a dummy user")
		mux.HandleFunc("POST /api/dev/login", srv.handleDevLogin)
		mux.HandleFunc("GET /api/dev/login", srv.handleDevLogin)
	}

	return mux
}

func newServer(st *store.Store, cfg Config) *Server {
	return &Server{
		store:            st,
		googleClientID:   cfg.GoogleClientID,
		useSecureCookies: !cfg.IsDevelopment,
		sessionTTL:       cfg.SessionTTL,
	}
}
