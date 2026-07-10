package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}

	port := getenv("PORT", "8080")
	dbPath := getenv("DB_PATH", "./mybudget.db")
	secureCookies := getenv("SECURE_COOKIES", "false") == "true"
	sessionTTL := parseTTL(getenv("SESSION_TTL", "720h"))
	devAuth := getenv("DEV_AUTH", "false") == "true"

	// The dummy-login shortcut is a full auth bypass; never let it run in a production
	// configuration. SECURE_COOKIES=true is our signal for "this is production".
	if devAuth && secureCookies {
		log.Fatal("DEV_AUTH must not be enabled with SECURE_COOKIES=true")
	}

	store, err := OpenStore(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	srv := &Server{
		store:          store,
		googleClientID: clientID,
		secureCookies:  secureCookies,
		sessionTTL:     sessionTTL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/google", srv.handleGoogleLogin)
	mux.HandleFunc("GET /api/me", srv.handleMe)
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

	// Dev-only dummy-login shortcut (see ADR-0010), enabled with DEV_AUTH=true so that
	// local and automated testing can reach authenticated screens without Google sign-in.
	if devAuth {
		log.Println("⚠  DEV AUTH ENABLED — /api/dev/login mints a session for a dummy user")
		mux.HandleFunc("POST /api/dev/login", srv.handleDevLogin)
		mux.HandleFunc("GET /api/dev/login", srv.handleDevLogin)
	}

	addr := ":" + port
	log.Printf("MyBudget API listening on %s (db=%s, secureCookies=%v)", addr, dbPath, secureCookies)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func parseTTL(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("invalid SESSION_TTL %q: %v", s, err)
	}

	return d
}
