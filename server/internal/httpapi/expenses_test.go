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
		name            string
		request         expenseRequest
		wantDescription string
		wantError       bool
	}{
		{
			name:            "normalizes valid request",
			request:         expenseRequest{CategoryID: 1, AmountMinor: 1599, Description: "  Burgers  ", SpentAt: "2026-07-11"},
			wantDescription: "Burgers",
		},
		{
			name:            "accepts leap day with blank description",
			request:         expenseRequest{CategoryID: 1, AmountMinor: 1, SpentAt: "2024-02-29"},
			wantDescription: "",
		},
		{
			name:            "accepts maximum length multibyte description",
			request:         expenseRequest{CategoryID: 1, AmountMinor: 1, Description: strings.Repeat("é", maxDescriptionLen), SpentAt: "2026-07-11"},
			wantDescription: strings.Repeat("é", maxDescriptionLen),
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
			name:      "rejects long description",
			request:   expenseRequest{CategoryID: 1, AmountMinor: 1, Description: strings.Repeat("🍔", maxDescriptionLen+1), SpentAt: "2026-07-11"},
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

			if !test.wantError && test.request.Description != test.wantDescription {
				t.Errorf("Description = %q, want %q", test.request.Description, test.wantDescription)
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
		bytes.NewBufferString(`{"categoryId":`+strconv.FormatInt(firstCategory.ID, 10)+`,"amountMinor":1599,"description":"  Burgers  ","spentAt":"2026-07-11"}`),
	)
	createRequest.Header.Set("Origin", testAppOrigin)
	createRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	createBody := createResponse.Body.Bytes()
	var created store.Expense
	if err := json.Unmarshal(createBody, &created); err != nil {
		t.Fatalf("decode created expense: %v", err)
	}

	if created.CategoryID != firstCategory.ID || created.AmountMinor != 1599 || created.Description != "Burgers" {
		t.Errorf("created expense = %#v, want normalized fields", created)
	}

	var createFields map[string]json.RawMessage
	if err := json.Unmarshal(createBody, &createFields); err != nil {
		t.Fatalf("decode created expense fields: %v", err)
	}

	if _, ok := createFields["description"]; !ok {
		t.Error("created expense response is missing description")
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

	updateRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/expenses/"+strconv.FormatInt(created.ID, 10),
		bytes.NewBufferString(`{"categoryId":`+strconv.FormatInt(firstCategory.ID, 10)+`,"amountMinor":1899,"description":"  Dinner  ","spentAt":"2026-07-10"}`),
	)
	updateRequest.Header.Set("Origin", testAppOrigin)
	updateRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	updateResponse := httptest.NewRecorder()
	handler.ServeHTTP(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResponse.Code, http.StatusOK)
	}

	var updated store.Expense
	if err := json.NewDecoder(updateResponse.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated expense: %v", err)
	}

	if updated.AmountMinor != 1899 || updated.Description != "Dinner" || updated.SpentAt != "2026-07-10" {
		t.Errorf("updated expense = %#v, want replacement fields", updated)
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
