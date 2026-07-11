package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// maxNameLen bounds user-provided names (budgets, categories) to keep inputs sane.
const maxNameLen = 100

type budgetRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Archived bool   `json:"archived"`
}

// normalizeAndValidate trims and validates a budget payload in place.
func (req *budgetRequest) normalizeAndValidate() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))

	if req.Name == "" || len(req.Name) > maxNameLen {
		return fmt.Errorf("name must be 1-%d characters", maxNameLen)
	}

	if !matchesCurrencyCodeFormat(req.Currency) {
		return fmt.Errorf("currency must be a 3-letter code")
	}

	return nil
}

// matchesCurrencyCodeFormat does a shape check (three A-Z letters); it is not a full ISO-4217 table.
func matchesCurrencyCodeFormat(code string) bool {
	if len(code) != 3 {
		return false
	}

	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return false
		}
	}

	return true
}

func (srv *Server) handleListBudgets(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgets, err := srv.store.ListBudgets(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "could not list budgets", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, budgets)
}

func (srv *Server) handleCreateBudget(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	var req budgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budget, err := srv.store.CreateBudget(r.Context(), user.ID, req.Name, req.Currency)
	if err != nil {
		http.Error(w, "could not create budget", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, budget)
}

func (srv *Server) handleGetBudget(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	budget, err := srv.store.GetBudget(r.Context(), user.ID, id)
	if err != nil {
		http.Error(w, "could not load budget", http.StatusInternalServerError)
		return
	}

	if budget == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, budget)
}

func (srv *Server) handleUpdateBudget(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var req budgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budget, err := srv.store.UpdateBudget(r.Context(), user.ID, id, req.Name, req.Currency, req.Archived)
	if err != nil {
		http.Error(w, "could not update budget", http.StatusInternalServerError)
		return
	}

	if budget == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, budget)
}

func (srv *Server) handleDeleteBudget(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	id, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	deleted, err := srv.store.DeleteBudget(r.Context(), user.ID, id)
	if err != nil {
		http.Error(w, "could not delete budget", http.StatusInternalServerError)
		return
	}

	if !deleted {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
