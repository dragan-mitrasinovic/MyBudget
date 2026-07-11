package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

func TestGetCurrentUserRenewsSessionAndCookieOncePerInterval(t *testing.T) {
	const sessionTTL = 30 * 24 * time.Hour
	st, db := openAuthTestStore(t)
	user, err := st.UpsertUser(
		context.Background(),
		"auth-test-user",
		"auth@example.com",
		"Auth Test User",
		"",
	)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	sessionID, err := st.CreateSession(context.Background(), user.ID, sessionTTL)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	oldExpiresAt := time.Now().UTC().Add(sessionTTL - 25*time.Hour).Format(time.RFC3339)
	if _, err := db.Exec("UPDATE sessions SET expires_at = ? WHERE id = ?", oldExpiresAt, sessionID); err != nil {
		t.Fatalf("age session: %v", err)
	}

	handler := New(st, Config{
		SessionTTL:    sessionTTL,
		IsDevelopment: true,
		AppOrigin:     "http://localhost:4200",
	})

	firstResponse := requestCurrentUser(t, handler, sessionID)
	if firstResponse.Code != http.StatusOK {
		t.Fatalf("first response status = %d, want %d", firstResponse.Code, http.StatusOK)
	}

	renewedCookie := findSessionCookie(t, firstResponse.Result().Cookies())
	if renewedCookie.Value != sessionID {
		t.Fatalf("renewed cookie value = %q, want %q", renewedCookie.Value, sessionID)
	}

	if renewedCookie.MaxAge != int(sessionTTL.Seconds()) {
		t.Fatalf("renewed cookie Max-Age = %d, want %d", renewedCookie.MaxAge, int(sessionTTL.Seconds()))
	}

	secondResponse := requestCurrentUser(t, handler, sessionID)
	if secondResponse.Code != http.StatusOK {
		t.Fatalf("second response status = %d, want %d", secondResponse.Code, http.StatusOK)
	}

	if cookies := secondResponse.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("second response cookies = %#v, want no renewal", cookies)
	}
}

func TestGetCurrentUserWithoutSessionIsUnauthorized(t *testing.T) {
	_, handler := openHTTPTestServer(t, false)
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestLogoutClearsAndInvalidatesSession(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	_, sessionID := createHTTPTestUserSession(t, st, "logout-user")

	requestLogout := func(sessionID string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		request.Header.Set("Origin", testAppOrigin)
		if sessionID != "" {
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	logoutResponse := requestLogout(sessionID)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", logoutResponse.Code, http.StatusNoContent)
	}

	clearedCookie := findSessionCookie(t, logoutResponse.Result().Cookies())
	if clearedCookie.Value != "" || clearedCookie.MaxAge != -1 || !clearedCookie.HttpOnly || !clearedCookie.Secure {
		t.Errorf("cleared cookie = %#v, want expired secure HttpOnly cookie", clearedCookie)
	}

	currentUserResponse := requestCurrentUser(t, handler, sessionID)
	if currentUserResponse.Code != http.StatusUnauthorized {
		t.Errorf("current-user status after logout = %d, want %d", currentUserResponse.Code, http.StatusUnauthorized)
	}

	idempotentResponse := requestLogout("")
	if idempotentResponse.Code != http.StatusNoContent {
		t.Errorf("logout without cookie status = %d, want %d", idempotentResponse.Code, http.StatusNoContent)
	}
}

func TestDevLoginCreatesUsableDevelopmentSession(t *testing.T) {
	_, handler := openHTTPTestServer(t, true)
	request := httptest.NewRequest(http.MethodGet, "/api/dev/login", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("dev login status = %d, want %d", response.Code, http.StatusOK)
	}

	sessionCookie := findSessionCookie(t, response.Result().Cookies())
	if sessionCookie.Value == "" || !sessionCookie.HttpOnly || sessionCookie.Secure || sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("development session cookie = %#v, want non-Secure HttpOnly Lax cookie", sessionCookie)
	}

	currentUserResponse := requestCurrentUser(t, handler, sessionCookie.Value)
	if currentUserResponse.Code != http.StatusOK {
		t.Errorf("current-user status = %d, want %d", currentUserResponse.Code, http.StatusOK)
	}
}

func openAuthTestStore(t *testing.T) (*store.Store, *sql.DB) {
	t.Helper()

	path := t.TempDir() + "/test.db"
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Fatalf("ping test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return st, db
}

func requestCurrentUser(t *testing.T, handler http.Handler, sessionID string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

func findSessionCookie(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == sessionCookieName {
			return cookie
		}
	}

	t.Fatal("session cookie not found")
	return nil
}
