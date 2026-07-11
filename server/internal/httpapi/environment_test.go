package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnvironmentMode(t *testing.T) {
	tests := []struct {
		name             string
		isDevelopment    bool
		wantSecureCookie bool
		wantDevRouteCode int
	}{
		{
			name:             "development",
			isDevelopment:    true,
			wantSecureCookie: false,
			wantDevRouteCode: http.StatusMethodNotAllowed,
		},
		{
			name:             "production",
			isDevelopment:    false,
			wantSecureCookie: true,
			wantDevRouteCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{IsDevelopment: tt.isDevelopment}
			srv := newServer(nil, cfg)

			cookieResponse := httptest.NewRecorder()
			srv.setSessionCookie(cookieResponse, "test-session")
			cookies := cookieResponse.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("got %d cookies, want 1", len(cookies))
			}

			if cookies[0].Secure != tt.wantSecureCookie {
				t.Errorf("Secure = %v, want %v", cookies[0].Secure, tt.wantSecureCookie)
			}

			request := httptest.NewRequest(http.MethodPut, "/api/dev/login", nil)
			response := httptest.NewRecorder()
			New(nil, cfg).ServeHTTP(response, request)
			if response.Code != tt.wantDevRouteCode {
				t.Errorf("dev route status = %d, want %d", response.Code, tt.wantDevRouteCode)
			}
		})
	}
}
