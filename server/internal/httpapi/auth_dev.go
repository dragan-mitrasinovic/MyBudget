package httpapi

import "net/http"

// Dev-only fixed identity used by the dummy-login shortcut.
const (
	devUserSub     = "dev-dummy-user"
	devUserEmail   = "dev@mybudget.local"
	devUserName    = "Dev User"
	devUserPicture = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='72' height='72'%3E%3Crect width='72' height='72' fill='%237e57c2'/%3E%3C/svg%3E"
)

// handleDevLogin upserts the fixed dummy user and starts a real session for it,
// mirroring handleGoogleLogin but skipping Google verification entirely.
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
