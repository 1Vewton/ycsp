package compute_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/compute"
)

// successBody 是数据处理层成功应答的样本，字段名与 Python 层保持一致。
const successBody = `{
  "service": "ycsp-compute",
  "status": "ok",
  "code": "prediction_ready",
  "prediction": {
    "formula": "NaCl",
    "family": "ionic",
    "candidates": [{
      "prototype": "rock_salt",
      "coordination_number": 6,
      "rule": {"code": "rule_radius_ratio", "params": {"ratio": 0.564}},
      "caveats": [],
      "cell": {
        "edge_a_angstrom": 5.66, "edge_b_angstrom": 5.66, "edge_c_angstrom": 5.66,
        "angle_alpha_degree": 90, "angle_beta_degree": 90, "angle_gamma_degree": 90
      },
      "atoms": [{"symbol": "Na", "fractional_coordinates": [0, 0, 0]}]
    }]
  }
}`

func newFakeCompute(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestPredictSendsFormulaAndParsesTheResult(t *testing.T) {
	var receivedBody []byte
	server := newFakeCompute(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		if request.URL.Path != "/predict" {
			t.Errorf("path = %s, want /predict", request.URL.Path)
		}
		receivedBody, _ = io.ReadAll(request.Body)
		_, _ = w.Write([]byte(successBody))
	})

	prediction, err := compute.NewClient(server.URL).Predict(context.Background(), "NaCl")
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}

	var requestPayload struct {
		Formula string `json:"formula"`
	}
	if err := json.Unmarshal(receivedBody, &requestPayload); err != nil {
		t.Fatalf("request body is not JSON: %v", err)
	}
	if requestPayload.Formula != "NaCl" {
		t.Errorf("sent formula = %q, want %q", requestPayload.Formula, "NaCl")
	}

	if prediction.Family != "ionic" || len(prediction.Candidates) != 1 {
		t.Fatalf("prediction = %+v", prediction)
	}
	candidate := prediction.Candidates[0]
	if candidate.Prototype != "rock_salt" || candidate.CoordinationNumber != 6 {
		t.Errorf("candidate = %+v", candidate)
	}
	if candidate.Cell.EdgeAAngstrom != 5.66 {
		t.Errorf("edge_a = %v, want 5.66", candidate.Cell.EdgeAAngstrom)
	}

	// rule / caveats / atoms 保留原始 JSON，落库时不用重新编码。
	var rule struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(candidate.Rule, &rule); err != nil || rule.Code != "rule_radius_ratio" {
		t.Errorf("rule = %s (err=%v)", candidate.Rule, err)
	}
	if string(candidate.Caveats) != "[]" {
		t.Errorf("caveats = %s, want []", candidate.Caveats)
	}
	if !strings.Contains(string(candidate.Atoms), `"symbol": "Na"`) {
		t.Errorf("atoms = %s", candidate.Atoms)
	}
}

func TestPredictPassesUpstreamRejectionThrough(t *testing.T) {
	server := newFakeCompute(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(
			`{"service":"ycsp-compute","status":"error","code":"formula_charge_unbalanced",` +
				`"params":{"formula":"NaCl2"}}`,
		))
	})

	_, err := compute.NewClient(server.URL).Predict(context.Background(), "NaCl2")

	var predictionError *compute.PredictionError
	if !errors.As(err, &predictionError) {
		t.Fatalf("error = %v, want PredictionError", err)
	}
	code, params := compute.Describe(err)
	if code != "formula_charge_unbalanced" {
		t.Errorf("code = %q, want the upstream code", code)
	}
	if params["formula"] != "NaCl2" {
		t.Errorf("params = %v", params)
	}
}

func TestPredictReportsUnexpectedStatusCode(t *testing.T) {
	server := newFakeCompute(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := compute.NewClient(server.URL).Predict(context.Background(), "NaCl")

	code, params := compute.Describe(err)
	if code != codes.ComputeBadStatus {
		t.Errorf("code = %q, want %q", code, codes.ComputeBadStatus)
	}
	if params["status_code"] != http.StatusInternalServerError {
		t.Errorf("params = %v", params)
	}
}

func TestPredictReportsInvalidJSON(t *testing.T) {
	server := newFakeCompute(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json at all"))
	})

	_, err := compute.NewClient(server.URL).Predict(context.Background(), "NaCl")

	if code, _ := compute.Describe(err); code != codes.ComputeInvalidJSON {
		t.Errorf("code = %q, want %q", code, codes.ComputeInvalidJSON)
	}
}

func TestPredictReportsEmptyResult(t *testing.T) {
	server := newFakeCompute(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok","prediction":{"formula":"NaCl","candidates":[]}}`))
	})

	_, err := compute.NewClient(server.URL).Predict(context.Background(), "NaCl")

	if code, _ := compute.Describe(err); code != codes.ComputeNoResult {
		t.Errorf("code = %q, want %q", code, codes.ComputeNoResult)
	}
}

func TestPredictReportsUnreachableLayer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	unreachableURL := server.URL
	server.Close()

	_, err := compute.NewClient(unreachableURL).Predict(context.Background(), "NaCl")

	code, params := compute.Describe(err)
	if code != codes.ComputeUnreachable {
		t.Errorf("code = %q, want %q", code, codes.ComputeUnreachable)
	}
	if !strings.HasSuffix(params["url"].(string), "/predict") {
		t.Errorf("params.url = %v", params["url"])
	}
}

func TestPredictReportsReadFailure(t *testing.T) {
	server := newFakeCompute(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "128")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"`))

		if hijacker, ok := w.(http.Hijacker); ok {
			connection, _, err := hijacker.Hijack()
			if err == nil {
				_ = connection.Close()
			}
		}
	})

	_, err := compute.NewClient(server.URL).Predict(context.Background(), "NaCl")

	if code, _ := compute.Describe(err); code != codes.ComputeReadFailed {
		t.Errorf("code = %q, want %q", code, codes.ComputeReadFailed)
	}
}
