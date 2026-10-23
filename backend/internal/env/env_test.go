package env_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/env"
)

func TestGetUsesFallbackWhenVariableIsEmpty(t *testing.T) {
	t.Setenv("YCSP_TEST_GET", "")

	if got := env.Get("YCSP_TEST_GET", "fallback"); got != "fallback" {
		t.Fatalf("Get = %q, want %q", got, "fallback")
	}
}

func TestGetPrefersTheEnvironmentValue(t *testing.T) {
	t.Setenv("YCSP_TEST_GET", "custom")

	if got := env.Get("YCSP_TEST_GET", "fallback"); got != "custom" {
		t.Fatalf("Get = %q, want %q", got, "custom")
	}
}

func TestGetIntParsesTheValue(t *testing.T) {
	t.Setenv("YCSP_TEST_PORT", "9000")

	port, err := env.GetInt("YCSP_TEST_PORT", 8000)
	if err != nil {
		t.Fatalf("GetInt: %v", err)
	}
	if port != 9000 {
		t.Fatalf("GetInt = %d, want 9000", port)
	}
}

func TestGetIntUsesFallback(t *testing.T) {
	t.Setenv("YCSP_TEST_PORT", "")

	port, err := env.GetInt("YCSP_TEST_PORT", 8000)
	if err != nil {
		t.Fatalf("GetInt: %v", err)
	}
	if port != 8000 {
		t.Fatalf("GetInt = %d, want 8000", port)
	}
}

func TestGetIntReportsInvalidValue(t *testing.T) {
	t.Setenv("YCSP_TEST_PORT", "not-a-number")

	_, err := env.GetInt("YCSP_TEST_PORT", 8000)
	if err == nil {
		t.Fatal("expected an error for a non-numeric port")
	}
	if !strings.Contains(err.Error(), "YCSP_TEST_PORT") {
		t.Errorf("error %q should name the offending variable", err)
	}
}

func TestLoadReadsEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("YCSP_TEST_LOADED=from-file\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("YCSP_TEST_LOADED") })

	env.Load(path)

	if got := env.Get("YCSP_TEST_LOADED", ""); got != "from-file" {
		t.Fatalf("Get = %q, want %q", got, "from-file")
	}
}

func TestLoadMissingFileDoesNotAbort(t *testing.T) {
	// 读不到 .env 只记一行日志：服务仍要用系统环境变量与默认值跑起来。
	env.Load(filepath.Join(t.TempDir(), "missing.env"))
}
