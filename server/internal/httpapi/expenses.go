package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

// maxNoteLen bounds the free-text note on an expense.
const maxNoteLen = 500

type expenseRequest struct {
	CategoryID  int64  `json:"categoryId"`
	AmountMinor int64  `json:"amountMinor"`
	Note        string `json:"note"`
	SpentAt     string `json:"spentAt"`
}

// normalizeAndValidate trims and validates an expense payload in place.
func (req *expenseRequest) normalizeAndValidate() error {
	req.Note = strings.TrimSpace(req.Note)

	if req.CategoryID <= 0 {
		return fmt.Errorf("categoryId is required")
	}

	if req.AmountMinor <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if len(req.Note) > maxNoteLen {
		return fmt.Errorf("note must be at most %d characters", maxNoteLen)
	}

	if !isValidISODate(req.SpentAt) {
		return fmt.Errorf("spentAt must be a valid date (YYYY-MM-DD)")
	}

	return nil
}

// isValidISODate reports whether s is a valid calendar date in YYYY-MM-DD form.
func isValidISODate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func (srv *Server) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgetID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	filter := store.ExpenseFilter{
		From: r.URL.Query().Get("from"),
		To:   r.URL.Query().Get("to"),
	}

	if raw := r.URL.Query().Get("categoryId"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid categoryId", http.StatusBadRequest)
			return
		}

		filter.CategoryID = &id
	}

	expenses, err := srv.store.ListExpenses(r.Context(), user.ID, budgetID, filter)
	if err != nil {
		writeExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, expenses)
}

func (srv *Server) handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	budgetID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var req expenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expense, err := srv.store.CreateExpense(
		r.Context(), user.ID, budgetID, req.CategoryID, req.AmountMinor, req.Note, req.SpentAt,
	)
	if err != nil {
		writeExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, expense)
}

func (srv *Server) handleUpdateExpense(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	expenseID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	var req expenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := req.normalizeAndValidate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expense, err := srv.store.UpdateExpense(
		r.Context(), user.ID, expenseID, req.CategoryID, req.AmountMinor, req.Note, req.SpentAt,
	)
	if err != nil {
		writeExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, expense)
}

func (srv *Server) handleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	user := srv.requireAuthenticatedUser(w, r)
	if user == nil {
		return
	}

	expenseID, ok := parsePathID(w, r, "id")
	if !ok {
		return
	}

	if err := srv.store.DeleteExpense(r.Context(), user.ID, expenseID); err != nil {
		writeExpenseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeExpenseError maps the expense data-layer sentinel errors to HTTP responses.
func writeExpenseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, store.ErrInvalidCategory):
		http.Error(w, "category does not belong to this budget", http.StatusBadRequest)
	default:
		http.Error(w, "server error", http.StatusInternalServerError)
	}
}
