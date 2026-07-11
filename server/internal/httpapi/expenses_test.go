package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/dragan-mitrasinovic/MyBudget/server/internal/store"
)

func TestExpenseRequestNormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   expenseRequest
		wantNote  string
		wantError bool
	}{
		{
			name:     "normalizes valid request",
			request:  expenseRequest{CategoryID: 1, AmountMinor: 1599, Note: "  Lunch  ", SpentAt: "2026-07-11"},
			wantNote: "Lunch",
		},
		{
			name:     "accepts leap day",
			request:  expenseRequest{CategoryID: 1, AmountMinor: 1, SpentAt: "2024-02-29"},
			wantNote: "",
		},
		{
			name:      "rejects missing category",
			request:   expenseRequest{AmountMinor: 1, SpentAt: "2026-07-11"},
			wantError: true,
		},
		{
			name:      "rejects non-positive amount",
			request:   expenseRequest{CategoryID: 1, SpentAt: "2026-07-11"},
			wantError: true,
		},
		{
			name:      "rejects long note",
			request:   expenseRequest{CategoryID: 1, AmountMinor: 1, Note: strings.Repeat("a", maxNoteLen+1), SpentAt: "2026-07-11"},
			wantError: true,
		},
		{
			name:      "rejects impossible date",
			request:   expenseRequest{CategoryID: 1, AmountMinor: 1, SpentAt: "2025-02-29"},
			wantError: true,
		},
		{
			name:      "rejects timestamp",
			request:   expenseRequest{CategoryID: 1, AmountMinor: 1, SpentAt: "2026-07-11T12:00:00Z"},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.normalizeAndValidate()
			if (err != nil) != test.wantError {
				t.Fatalf("normalizeAndValidate() error = %v, wantError %v", err, test.wantError)
			}

			if !test.wantError && test.request.Note != test.wantNote {
				t.Errorf("Note = %q, want %q", test.request.Note, test.wantNote)
			}
		})
	}
}

func TestExpenseEndpointsCreateListAndRejectWrongBudgetCategory(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	user, sessionID := createHTTPTestUserSession(t, st, "expense-owner")
	firstBudget, err := st.CreateBudget(t.Context(), user.ID, "First", "EUR")
	if err != nil {
		t.Fatalf("create first budget: %v", err)
	}

	secondBudget, err := st.CreateBudget(t.Context(), user.ID, "Second", "EUR")
	if err != nil {
		t.Fatalf("create second budget: %v", err)
	}

	firstCategory, err := st.CreateCategory(t.Context(), user.ID, firstBudget.ID, "Food")
	if err != nil {
		t.Fatalf("create first category: %v", err)
	}

	secondCategory, err := st.CreateCategory(t.Context(), user.ID, secondBudget.ID, "Travel")
	if err != nil {
		t.Fatalf("create second category: %v", err)
	}

	collectionPath := "/api/budgets/" + strconv.FormatInt(firstBudget.ID, 10) + "/expenses"
	createRequest := httptest.NewRequest(
		http.MethodPost,
		collectionPath,
		bytes.NewBufferString(`{"categoryId":`+strconv.FormatInt(firstCategory.ID, 10)+`,"amountMinor":1599,"note":"  Lunch  ","spentAt":"2026-07-11"}`),
	)
	createRequest.Header.Set("Origin", testAppOrigin)
	createRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	var created store.Expense
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created expense: %v", err)
	}

	if created.CategoryID != firstCategory.ID || created.AmountMinor != 1599 || created.Note != "Lunch" {
		t.Errorf("created expense = %#v, want normalized fields", created)
	}

	listRequest := httptest.NewRequest(http.MethodGet, collectionPath, nil)
	listRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResponse.Code, http.StatusOK)
	}

	var expenses []store.Expense
	if err := json.NewDecoder(listResponse.Body).Decode(&expenses); err != nil {
		t.Fatalf("decode expense list: %v", err)
	}

	if len(expenses) != 1 || expenses[0].ID != created.ID {
		t.Fatalf("listed expenses = %#v, want created expense", expenses)
	}

	wrongCategoryRequest := httptest.NewRequest(
		http.MethodPost,
		collectionPath,
		bytes.NewBufferString(`{"categoryId":`+strconv.FormatInt(secondCategory.ID, 10)+`,"amountMinor":100,"spentAt":"2026-07-11"}`),
	)
	wrongCategoryRequest.Header.Set("Origin", testAppOrigin)
	wrongCategoryRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	wrongCategoryResponse := httptest.NewRecorder()
	handler.ServeHTTP(wrongCategoryResponse, wrongCategoryRequest)

	if wrongCategoryResponse.Code != http.StatusBadRequest {
		t.Errorf("wrong-category status = %d, want %d", wrongCategoryResponse.Code, http.StatusBadRequest)
	}
}
