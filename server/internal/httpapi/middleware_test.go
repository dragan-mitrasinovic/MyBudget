package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAppOrigin(t *testing.T) {
	const appOrigin = "http://localhost:4200"
	tests := []struct {
		name       string
		origin     string
		wantStatus int
		wantCalled bool
	}{
		{name: "matching origin", origin: appOrigin, wantStatus: http.StatusNoContent, wantCalled: true},
		{name: "missing origin", wantStatus: http.StatusForbidden},
		{name: "different origin", origin: "https://example.com", wantStatus: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wasCalled := false
			srv := &Server{appOrigin: appOrigin}
			handler := srv.requireAppOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				wasCalled = true
				w.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request.Header.Set("Origin", test.origin)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, test.wantStatus)
			}

			if wasCalled != test.wantCalled {
				t.Errorf("handler called = %v, want %v", wasCalled, test.wantCalled)
			}
		})
	}
}
