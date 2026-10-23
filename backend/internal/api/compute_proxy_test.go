package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/codes"
)

// newFakeComputeServer 起一个假的“数据处理层”，测试结束自动关闭。
func newFakeComputeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestPingComputeForwardsUpstreamHealth(t *testing.T) {
	var requestedPath string
	server := newFakeComputeServer(t, func(w http.ResponseWriter, request *http.Request) {
		requestedPath = request.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(
			`{"service":"fake-compute","version":"9.9.9","status":"ok","code":"compute_healthy"}`,
		))
	})

	router := api.NewRouter(api.Options{ComputeBaseURL: server.URL})
	recorder, payload := performGET(t, router, "/api/ping-compute")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if requestedPath != "/health" {
		t.Errorf("upstream path = %q, want %q", requestedPath, "/health")
	}
	if payload["code"] != codes.ComputeHealthy {
		t.Errorf("code = %v, want %q", payload["code"], codes.ComputeHealthy)
	}

	computePayload, ok := payload["compute"].(map[string]any)
	if !ok {
		t.Fatalf("compute = %T, want an object", payload["compute"])
	}
	if computePayload["version"] != "9.9.9" {
		t.Errorf("compute.version = %v, want %q", computePayload["version"], "9.9.9")
	}
}

func TestPingComputeReportsBadStatusCode(t *testing.T) {
	server := newFakeComputeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	router := api.NewRouter(api.Options{ComputeBaseURL: server.URL})
	recorder, payload := performGET(t, router, "/api/ping-compute")

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if payload["code"] != codes.ComputeBadStatus {
		t.Errorf("code = %v, want %q", payload["code"], codes.ComputeBadStatus)
	}

	params, ok := payload["params"].(map[string]any)
	if !ok {
		t.Fatalf("params = %T, want an object", payload["params"])
	}
	if params["status_code"] != float64(http.StatusInternalServerError) {
		t.Errorf("params.status_code = %v, want %d", params["status_code"], http.StatusInternalServerError)
	}
}

func TestPingComputeReportsInvalidJSON(t *testing.T) {
	server := newFakeComputeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json at all"))
	})

	router := api.NewRouter(api.Options{ComputeBaseURL: server.URL})
	recorder, payload := performGET(t, router, "/api/ping-compute")

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if payload["code"] != codes.ComputeInvalidJSON {
		t.Errorf("code = %v, want %q", payload["code"], codes.ComputeInvalidJSON)
	}
}

func TestPingComputeReportsUnreachableLayer(t *testing.T) {
	// 先起再关，拿到一个立刻拒绝连接的地址，模拟数据处理层没启动。
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	unreachableURL := server.URL
	server.Close()

	router := api.NewRouter(api.Options{ComputeBaseURL: unreachableURL})
	recorder, payload := performGET(t, router, "/api/ping-compute")

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if payload["code"] != codes.ComputeUnreachable {
		t.Errorf("code = %v, want %q", payload["code"], codes.ComputeUnreachable)
	}

	params, ok := payload["params"].(map[string]any)
	if !ok {
		t.Fatalf("params = %T, want an object", payload["params"])
	}
	if params["url"] != unreachableURL+"/health" {
		t.Errorf("params.url = %v, want %q", params["url"], unreachableURL+"/health")
	}
}

func TestPingComputeReportsReadFailure(t *testing.T) {
	// 声明比实际更长的 Content-Length 再断开连接，让后端读应答时出错。
	server := newFakeComputeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "128")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service":"fake-compute"`))

		if hijacker, ok := w.(http.Hijacker); ok {
			connection, _, err := hijacker.Hijack()
			if err == nil {
				_ = connection.Close()
			}
		}
	})

	router := api.NewRouter(api.Options{ComputeBaseURL: server.URL})
	recorder, payload := performGET(t, router, "/api/ping-compute")

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if payload["code"] != codes.ComputeReadFailed {
		t.Errorf("code = %v, want %q", payload["code"], codes.ComputeReadFailed)
	}
}
