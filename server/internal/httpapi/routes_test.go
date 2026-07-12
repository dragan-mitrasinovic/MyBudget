package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

const testAppOrigin = "http://localhost:4200"

type testRoute struct {
	name   string
	method string
	path   string
}

func TestProtectedRoutesRequireAuthentication(t *testing.T) {
	_, handler := openHTTPTestServer(t, false)
	routes := []testRoute{
		{name: "current user", method: http.MethodGet, path: "/api/me"},
		{name: "list budgets", method: http.MethodGet, path: "/api/budgets"},
		{name: "create budget", method: http.MethodPost, path: "/api/budgets"},
		{name: "get budget", method: http.MethodGet, path: "/api/budgets/1"},
		{name: "update budget", method: http.MethodPatch, path: "/api/budgets/1"},
		{name: "delete budget", method: http.MethodDelete, path: "/api/budgets/1"},
		{name: "get dashboard layout", method: http.MethodGet, path: "/api/budgets/1/dashboard"},
		{name: "update dashboard layout", method: http.MethodPut, path: "/api/budgets/1/dashboard"},
		{name: "list categories", method: http.MethodGet, path: "/api/budgets/1/categories"},
		{name: "create category", method: http.MethodPost, path: "/api/budgets/1/categories"},
		{name: "update category", method: http.MethodPatch, path: "/api/categories/1"},
		{name: "delete category", method: http.MethodDelete, path: "/api/categories/1"},
		{name: "list expenses", method: http.MethodGet, path: "/api/budgets/1/expenses"},
		{name: "create expense", method: http.MethodPost, path: "/api/budgets/1/expenses"},
		{name: "update expense", method: http.MethodPatch, path: "/api/expenses/1"},
		{name: "delete expense", method: http.MethodDelete, path: "/api/expenses/1"},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, nil)
			if route.method != http.MethodGet {
				request.Header.Set("Origin", testAppOrigin)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestStateChangingRoutesRequireAppOrigin(t *testing.T) {
	_, handler := openHTTPTestServer(t, true)
	routes := []testRoute{
		{name: "google login", method: http.MethodPost, path: "/api/auth/google"},
		{name: "logout", method: http.MethodPost, path: "/api/auth/logout"},
		{name: "dev login", method: http.MethodPost, path: "/api/dev/login"},
		{name: "create budget", method: http.MethodPost, path: "/api/budgets"},
		{name: "update budget", method: http.MethodPatch, path: "/api/budgets/1"},
		{name: "delete budget", method: http.MethodDelete, path: "/api/budgets/1"},
		{name: "update dashboard layout", method: http.MethodPut, path: "/api/budgets/1/dashboard"},
		{name: "create category", method: http.MethodPost, path: "/api/budgets/1/categories"},
		{name: "update category", method: http.MethodPatch, path: "/api/categories/1"},
		{name: "delete category", method: http.MethodDelete, path: "/api/categories/1"},
		{name: "create expense", method: http.MethodPost, path: "/api/budgets/1/expenses"},
		{name: "update expense", method: http.MethodPatch, path: "/api/expenses/1"},
		{name: "delete expense", method: http.MethodDelete, path: "/api/expenses/1"},
	}
	origins := []struct {
		name  string
		value string
	}{
		{name: "missing"},
		{name: "different", value: "https://example.com"},
	}

	for _, route := range routes {
		for _, origin := range origins {
			t.Run(route.name+"/"+origin.name, func(t *testing.T) {
				request := httptest.NewRequest(route.method, route.path, nil)
				request.Header.Set("Origin", origin.value)
				response := httptest.NewRecorder()

				handler.ServeHTTP(response, request)

				if response.Code != http.StatusForbidden {
					t.Errorf("status = %d, want %d", response.Code, http.StatusForbidden)
				}
			})
		}
	}
}

func openHTTPTestServer(t *testing.T, isDevelopment bool) (*store.Store, http.Handler) {
	t.Helper()

	st, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	handler := New(st, Config{
		GoogleClientID: "test-client-id",
		SessionTTL:     30 * 24 * time.Hour,
		IsDevelopment:  isDevelopment,
		AppOrigin:      testAppOrigin,
	})

	return st, handler
}

func createHTTPTestUserSession(t *testing.T, st *store.Store, sub string) (*store.User, string) {
	t.Helper()

	user, err := st.UpsertUser(context.Background(), sub, sub+"@example.com", sub, "")
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	sessionID, err := st.CreateSession(context.Background(), user.ID, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("create test session: %v", err)
	}

	return user, sessionID
}
