package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/codes"
)

// performGET 发一个 GET 请求，并把应答解析成通用 map，方便断言 code / params。
func performGET(t *testing.T, router http.Handler, path string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not JSON: %v (body=%q)", err, recorder.Body.String())
	}
	return recorder, payload
}

func TestHealthReturnsLanguageNeutralCode(t *testing.T) {
	router := api.NewRouter(api.Options{ComputeBaseURL: "http://127.0.0.1:1"})

	recorder, payload := performGET(t, router, "/api/health")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if payload["service"] != api.ServiceName {
		t.Errorf("service = %v, want %q", payload["service"], api.ServiceName)
	}
	if payload["status"] != "ok" {
		t.Errorf("status field = %v, want %q", payload["status"], "ok")
	}
	if payload["code"] != codes.BackendHealthy {
		t.Errorf("code = %v, want %q", payload["code"], codes.BackendHealthy)
	}
}

func TestUnknownRouteIsNotHandled(t *testing.T) {
	router := api.NewRouter(api.Options{ComputeBaseURL: "http://127.0.0.1:1"})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/unknown", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestRequestIDIsEchoedOrGenerated(t *testing.T) {
	router := api.NewRouter(api.Options{ComputeBaseURL: "http://127.0.0.1:1"})

	withClientID := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	withClientID.Header.Set("X-Request-ID", "web-test-123")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, withClientID)
	if got := recorder.Header().Get("X-Request-ID"); got != "web-test-123" {
		t.Errorf("echoed request id = %q, want web-test-123", got)
	}

	withoutID := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, withoutID)
	if got := recorder.Header().Get("X-Request-ID"); !strings.HasPrefix(got, "api-") {
		t.Errorf("generated request id = %q, want an api- prefix", got)
	}

	invalidID := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	invalidID.Header.Set("X-Request-ID", "web 123")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, invalidID)
	if got := recorder.Header().Get("X-Request-ID"); !strings.HasPrefix(got, "api-") {
		t.Errorf("request id for an invalid header = %q, want a generated api- id", got)
	}
}
