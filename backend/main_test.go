package main

import (
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/store"
)

// clearSettings 把配置项清空（空值等于未设置），让 readSettings 走默认值。
func clearSettings(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"YCSP_BACKEND_HOST",
		"YCSP_BACKEND_PORT",
		"YCSP_COMPUTE_BASE_URL",
		"YCSP_DB_DRIVER",
		"YCSP_DB_DSN",
	} {
		t.Setenv(key, "")
	}
}

func TestReadSettingsUsesDefaults(t *testing.T) {
	clearSettings(t)

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if settings.listenHost != "127.0.0.1" {
		t.Errorf("listenHost = %q, want %q", settings.listenHost, "127.0.0.1")
	}
	if settings.listenPort != 8000 {
		t.Errorf("listenPort = %d, want %d", settings.listenPort, 8000)
	}
	if settings.computeBaseURL != "http://127.0.0.1:8001" {
		t.Errorf("computeBaseURL = %q, want %q", settings.computeBaseURL, "http://127.0.0.1:8001")
	}
	if settings.database.Driver != store.DriverSQLite {
		t.Errorf("database driver = %q, want %q", settings.database.Driver, store.DriverSQLite)
	}
	if settings.database.DSN != "ycsp.db" {
		t.Errorf("database DSN = %q, want %q", settings.database.DSN, "ycsp.db")
	}
}

func TestReadSettingsPrefersTheEnvironment(t *testing.T) {
	t.Setenv("YCSP_BACKEND_HOST", "0.0.0.0")
	t.Setenv("YCSP_BACKEND_PORT", "9000")
	t.Setenv("YCSP_COMPUTE_BASE_URL", "http://compute.local:9001")
	t.Setenv("YCSP_DB_DRIVER", "postgres")
	t.Setenv("YCSP_DB_DSN", "host=db.local dbname=ycsp")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if settings.listenHost != "0.0.0.0" || settings.listenPort != 9000 {
		t.Errorf("listen = %s:%d, want 0.0.0.0:9000", settings.listenHost, settings.listenPort)
	}
	if settings.computeBaseURL != "http://compute.local:9001" {
		t.Errorf("computeBaseURL = %q", settings.computeBaseURL)
	}
	if settings.database.Driver != store.DriverPostgres {
		t.Errorf("database driver = %q, want %q", settings.database.Driver, store.DriverPostgres)
	}
	if settings.database.DSN != "host=db.local dbname=ycsp" {
		t.Errorf("database DSN = %q", settings.database.DSN)
	}
}

func TestReadSettingsRejectsInvalidPort(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_BACKEND_PORT", "8000x")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for a non-numeric port")
	}
}

func TestReadSettingsLeavesDSNEmptyForNonSQLite(t *testing.T) {
	// postgres / mysql 的连接串含地址与账号，不能猜：留空，让 store.Open 明确报错。
	clearSettings(t)
	t.Setenv("YCSP_DB_DRIVER", "postgres")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}
	if settings.database.DSN != "" {
		t.Errorf("database DSN = %q, want empty", settings.database.DSN)
	}
}
