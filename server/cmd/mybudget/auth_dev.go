package main

import "net/http"

// Dev-only fixed identity used by the dummy-login shortcut (enabled with
// DEV_AUTH=true; never in production — see main.go's guard and ADR-0010). The
// google_sub is a sentinel value that can never collide with a real Google subject.
const (
	devUserSub   = "dev-dummy-user"
	devUserEmail = "dev@mybudget.local"
	devUserName  = "Dev User"
	// Self-contained placeholder avatar (a solid deep-purple square, rendered as a
	// circle by the UI) so the profile image isn't a broken <img> in local testing.
	devUserPicture = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='72' height='72'%3E%3Crect width='72' height='72' fill='%237e57c2'/%3E%3C/svg%3E"
)

// handleDevLogin upserts the fixed dummy user and starts a real session for it,
// mirroring handleGoogleLogin but skipping Google verification entirely. It is only
// wired up when DEV_AUTH=true (see main.go).
func (srv *Server) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	user, err := srv.store.UpsertUser(r.Context(), devUserSub, devUserEmail, devUserName, devUserPicture)
	if err != nil {
		http.Error(w, "could not save dev user", http.StatusInternalServerError)
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
