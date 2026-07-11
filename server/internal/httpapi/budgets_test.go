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

func TestBudgetRequestNormalizeAndValidate(t *testing.T) {
	zeroTarget := int64(0)
	negativeTarget := int64(-1)

	tests := []struct {
		name         string
		request      budgetRequest
		wantName     string
		wantCurrency string
		wantError    bool
	}{
		{
			name:         "normalizes valid request",
			request:      budgetRequest{Name: "  Household  ", Currency: " eur ", TargetMinor: &zeroTarget},
			wantName:     "Household",
			wantCurrency: "EUR",
		},
		{
			name:         "accepts nil target",
			request:      budgetRequest{Name: "Household", Currency: "EUR"},
			wantName:     "Household",
			wantCurrency: "EUR",
		},
		{
			name:      "rejects empty name",
			request:   budgetRequest{Name: "  ", Currency: "EUR"},
			wantError: true,
		},
		{
			name:      "rejects long name",
			request:   budgetRequest{Name: strings.Repeat("a", maxNameLen+1), Currency: "EUR"},
			wantError: true,
		},
		{
			name:      "rejects short currency",
			request:   budgetRequest{Name: "Household", Currency: "EU"},
			wantError: true,
		},
		{
			name:      "rejects non-letter currency",
			request:   budgetRequest{Name: "Household", Currency: "E1R"},
			wantError: true,
		},
		{
			name:      "rejects negative target",
			request:   budgetRequest{Name: "Household", Currency: "EUR", TargetMinor: &negativeTarget},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.normalizeAndValidate()
			if (err != nil) != test.wantError {
				t.Fatalf("normalizeAndValidate() error = %v, wantError %v", err, test.wantError)
			}

			if test.wantError {
				return
			}

			if test.request.Name != test.wantName {
				t.Errorf("Name = %q, want %q", test.request.Name, test.wantName)
			}

			if test.request.Currency != test.wantCurrency {
				t.Errorf("Currency = %q, want %q", test.request.Currency, test.wantCurrency)
			}
		})
	}
}

func TestBudgetEndpointsCreateAndList(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	_, sessionID := createHTTPTestUserSession(t, st, "budget-owner")

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/budgets",
		bytes.NewBufferString(`{"name":"  Household  ","currency":" eur ","targetMinor":120000}`),
	)
	createRequest.Header.Set("Origin", testAppOrigin)
	createRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	var created store.Budget
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created budget: %v", err)
	}

	if created.Name != "Household" || created.Currency != "EUR" || created.TargetMinor == nil || *created.TargetMinor != 120000 {
		t.Errorf("created budget = %#v, want normalized fields", created)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/budgets", nil)
	listRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionID})
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResponse.Code, http.StatusOK)
	}

	var budgets []store.Budget
	if err := json.NewDecoder(listResponse.Body).Decode(&budgets); err != nil {
		t.Fatalf("decode budget list: %v", err)
	}

	if len(budgets) != 1 || budgets[0].ID != created.ID {
		t.Fatalf("listed budgets = %#v, want created budget", budgets)
	}
}

func TestGetBudgetDoesNotDiscloseOtherUsersBudget(t *testing.T) {
	st, handler := openHTTPTestServer(t, false)
	owner, _ := createHTTPTestUserSession(t, st, "budget-owner")
	_, otherSessionID := createHTTPTestUserSession(t, st, "other-user")

	budget, err := st.CreateBudget(t.Context(), owner.ID, "Household", "EUR", nil)
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/budgets/"+strconv.FormatInt(budget.ID, 10), nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: otherSessionID})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
