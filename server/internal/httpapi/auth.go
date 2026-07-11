package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
	"google.golang.org/api/idtoken"
)

const sessionCookieName = "session"

type googleLoginRequest struct {
	Credential string `json:"credential"`
}

// handleGoogleLogin verifies a Google ID token, upserts the user, and starts a session.
// Registration is implicit: a first-time user is created here.
func (srv *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req googleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Credential == "" {
		http.Error(w, "missing credential", http.StatusBadRequest)
		return
	}

	// Validate verifies the signature against Google's public keys and checks that
	// aud == our client ID, iss is Google, and the token has not expired.
	payload, err := idtoken.Validate(r.Context(), req.Credential, srv.googleClientID)
	if err != nil {
		http.Error(w, "invalid credential", http.StatusUnauthorized)
		return
	}

	if verified, _ := payload.Claims["email_verified"].(bool); !verified {
		http.Error(w, "email not verified", http.StatusUnauthorized)
		return
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	user, err := srv.store.UpsertUser(r.Context(), payload.Subject, email, name, picture)
	if err != nil {
		http.Error(w, "could not save user", http.StatusInternalServerError)
		return
	}

	sessionID, err := srv.store.CreateSession(r.Context(), user.ID, srv.sessionTTL)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}

	srv.setSessionCookie(w, sessionID)
	writeJSON(w, http.StatusOK, user)
}

// handleMe returns the signed-in user, or 401 if there is no valid session.
func (srv *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := srv.currentUser(r)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleLogout deletes the session and clears the cookie.
func (srv *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		_ = srv.store.DeleteSession(r.Context(), c.Value)
	}

	srv.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// currentUser resolves the session cookie to a user, returning (nil, nil) when there
// is no cookie or the session is invalid/expired.
func (srv *Server) currentUser(r *http.Request) (*store.User, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}

		return nil, err
	}

	return srv.store.GetSessionUser(r.Context(), c.Value)
}

func (srv *Server) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   srv.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(srv.sessionTTL.Seconds()),
	})
}

func (srv *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   srv.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
