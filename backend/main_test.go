package main

import (
	"net/netip"
	"testing"
	"time"

	"github.com/1Vewton/ycsp/backend/internal/store"
)

// clearSettings 把配置项清空（空值等于未设置），让 readSettings 走默认值。
func clearSettings(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"YCSP_ENV",
		"YCSP_BACKEND_HOST",
		"YCSP_BACKEND_PORT",
		"YCSP_COMPUTE_BASE_URL",
		"YCSP_DB_DRIVER",
		"YCSP_DB_DSN",
		"YCSP_DB_MAX_OPEN_CONNS",
		"YCSP_DB_MAX_IDLE_CONNS",
		"YCSP_DB_CONN_MAX_LIFETIME_SECONDS",
		"YCSP_DB_AUTO_MIGRATE",
		"YCSP_TASK_WORKERS",
		"YCSP_SESSION_DRIVER",
		"YCSP_REDIS_URL",
		"YCSP_SESSION_LIFETIME_SECONDS",
		"YCSP_COOKIE_SECURE",
		"YCSP_LOGIN_MAX_ATTEMPTS",
		"YCSP_LOGIN_IP_MAX_ATTEMPTS",
		"YCSP_LOGIN_BLOCK_SECONDS",
		"YCSP_TRUSTED_PROXIES",
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
	if settings.environment != environmentDevelopment {
		t.Errorf("environment = %q, want %q", settings.environment, environmentDevelopment)
	}
	if !settings.autoMigrate {
		t.Error("autoMigrate should default to true outside production")
	}
	if settings.taskWorkers != sqliteDefaultTaskWorkers {
		t.Errorf("taskWorkers = %d, want %d for sqlite", settings.taskWorkers, sqliteDefaultTaskWorkers)
	}
	if settings.database.Driver != store.DriverSQLite {
		t.Errorf("database driver = %q, want %q", settings.database.Driver, store.DriverSQLite)
	}
	if settings.database.DSN != "ycsp.db" {
		t.Errorf("database DSN = %q, want %q", settings.database.DSN, "ycsp.db")
	}
	if settings.database.MaxOpenConns != 0 || settings.database.MaxIdleConns != 0 {
		t.Errorf(
			"database pool = (%d, %d), want zero so the driver defaults apply",
			settings.database.MaxOpenConns, settings.database.MaxIdleConns,
		)
	}
	if settings.database.ConnMaxLifetime != 0 {
		t.Errorf("database lifetime = %v, want zero (driver default)", settings.database.ConnMaxLifetime)
	}
	if settings.sessions.driver != "redis" {
		t.Errorf("session driver = %q, want %q", settings.sessions.driver, "redis")
	}
	if settings.sessions.redisURL != "redis://127.0.0.1:6379/0" {
		t.Errorf("redis URL = %q", settings.sessions.redisURL)
	}
	if settings.sessions.lifetime != 7*24*time.Hour {
		t.Errorf("session lifetime = %v, want 7 days", settings.sessions.lifetime)
	}
	if settings.sessions.secureCookie {
		t.Error("cookie secure should default to false for local http development")
	}
	if settings.loginLimit.maxAttemptsPerAccount != 5 {
		t.Errorf("login attempts per account = %d, want 5", settings.loginLimit.maxAttemptsPerAccount)
	}
	if settings.loginLimit.maxAttemptsPerAddress != 20 {
		t.Errorf("login attempts per address = %d, want 20", settings.loginLimit.maxAttemptsPerAddress)
	}
	if settings.loginLimit.window != 15*time.Minute {
		t.Errorf("login block window = %v, want 15m", settings.loginLimit.window)
	}
}

func TestReadSettingsUsesProductionPostgresDefaults(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_ENV", environmentProduction)
	t.Setenv("YCSP_DB_DRIVER", store.DriverPostgres)
	t.Setenv("YCSP_DB_DSN", "host=db.local dbname=ycsp")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if settings.environment != environmentProduction {
		t.Errorf("environment = %q, want %q", settings.environment, environmentProduction)
	}
	if settings.autoMigrate {
		t.Error("production should not auto-migrate by default")
	}
	if settings.taskWorkers != serverDefaultTaskWorkers {
		t.Errorf("taskWorkers = %d, want %d for a server database", settings.taskWorkers, serverDefaultTaskWorkers)
	}
}

func TestReadSettingsRejectsSQLiteInProduction(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_ENV", environmentProduction)

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for sqlite in production")
	}
}

func TestReadSettingsPrefersTaskWorkersAndPool(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_DB_DRIVER", store.DriverPostgres)
	t.Setenv("YCSP_DB_DSN", "host=db.local dbname=ycsp")
	t.Setenv("YCSP_TASK_WORKERS", "8")
	t.Setenv("YCSP_DB_MAX_OPEN_CONNS", "30")
	t.Setenv("YCSP_DB_MAX_IDLE_CONNS", "15")
	t.Setenv("YCSP_DB_CONN_MAX_LIFETIME_SECONDS", "600")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if settings.taskWorkers != 8 {
		t.Errorf("taskWorkers = %d, want 8", settings.taskWorkers)
	}
	if settings.database.MaxOpenConns != 30 || settings.database.MaxIdleConns != 15 {
		t.Errorf(
			"database pool = (%d, %d), want (30, 15)",
			settings.database.MaxOpenConns, settings.database.MaxIdleConns,
		)
	}
	if settings.database.ConnMaxLifetime != 10*time.Minute {
		t.Errorf("database lifetime = %v, want 10m", settings.database.ConnMaxLifetime)
	}
}

func TestReadSettingsRejectsInvalidTaskWorkers(t *testing.T) {
	for _, workers := range []string{"-1", "17"} {
		clearSettings(t)
		t.Setenv("YCSP_TASK_WORKERS", workers)
		if _, err := readSettings(); err == nil {
			t.Errorf("expected an error for YCSP_TASK_WORKERS=%s", workers)
		}
	}
}

func TestReadSettingsRejectsInvalidEnvironment(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_ENV", "staging")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for an unknown environment")
	}
}

func TestReadSettingsRejectsInvalidAutoMigrate(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_DB_AUTO_MIGRATE", "yes")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for a non-boolean auto-migrate flag")
	}
}

func TestReadSettingsRejectsNegativePoolValues(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_DB_MAX_OPEN_CONNS", "-1")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for a negative pool size")
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

func TestReadSettingsPrefersSessionEnvironment(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_SESSION_DRIVER", "memory")
	t.Setenv("YCSP_REDIS_URL", "redis://cache.local:6380/2")
	t.Setenv("YCSP_SESSION_LIFETIME_SECONDS", "3600")
	t.Setenv("YCSP_COOKIE_SECURE", "true")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if settings.sessions.driver != "memory" {
		t.Errorf("session driver = %q, want memory", settings.sessions.driver)
	}
	if settings.sessions.redisURL != "redis://cache.local:6380/2" {
		t.Errorf("redis URL = %q", settings.sessions.redisURL)
	}
	if settings.sessions.lifetime != time.Hour {
		t.Errorf("session lifetime = %v, want 1h", settings.sessions.lifetime)
	}
	if !settings.sessions.secureCookie {
		t.Error("cookie secure should be true")
	}
}

func TestReadSettingsRejectsInvalidSessionLifetime(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_SESSION_LIFETIME_SECONDS", "one-week")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for a non-numeric session lifetime")
	}
}

func TestReadSettingsRejectsMisspelledCookieSecure(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_COOKIE_SECURE", "yes-please")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for a non-boolean cookie flag")
	}
}

func TestReadSettingsPrefersLoginLimitEnvironment(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_LOGIN_MAX_ATTEMPTS", "3")
	t.Setenv("YCSP_LOGIN_IP_MAX_ATTEMPTS", "9")
	t.Setenv("YCSP_LOGIN_BLOCK_SECONDS", "120")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if settings.loginLimit.maxAttemptsPerAccount != 3 {
		t.Errorf("account attempts = %d, want 3", settings.loginLimit.maxAttemptsPerAccount)
	}
	if settings.loginLimit.maxAttemptsPerAddress != 9 {
		t.Errorf("address attempts = %d, want 9", settings.loginLimit.maxAttemptsPerAddress)
	}
	if settings.loginLimit.window != 2*time.Minute {
		t.Errorf("window = %v, want 2m", settings.loginLimit.window)
	}
}

func TestReadSettingsRejectsInvalidLoginLimit(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_LOGIN_MAX_ATTEMPTS", "five")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for a non-numeric login limit")
	}
}

func TestReadSettingsParsesTrustedProxies(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_TRUSTED_PROXIES", "127.0.0.1, 10.0.0.0/8")

	settings, err := readSettings()
	if err != nil {
		t.Fatalf("readSettings: %v", err)
	}

	if len(settings.trustedProxies) != 2 {
		t.Fatalf("trusted proxy count = %d, want 2", len(settings.trustedProxies))
	}
	if !settings.trustedProxies[0].Contains(netip.MustParseAddr("127.0.0.1")) {
		t.Errorf("first trusted proxy = %v, want 127.0.0.1/32", settings.trustedProxies[0])
	}
	if !settings.trustedProxies[1].Contains(netip.MustParseAddr("10.4.5.6")) {
		t.Errorf("second trusted proxy = %v, want 10.0.0.0/8", settings.trustedProxies[1])
	}
}

func TestReadSettingsRejectsInvalidTrustedProxy(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_TRUSTED_PROXIES", "nginx.local")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for an invalid trusted proxy")
	}
}

func TestReadSettingsRejectsTrustingEveryAddress(t *testing.T) {
	clearSettings(t)
	t.Setenv("YCSP_TRUSTED_PROXIES", "0.0.0.0/0")

	if _, err := readSettings(); err == nil {
		t.Fatal("expected an error for 0.0.0.0/0")
	}
}

func TestNewSessionBackendSupportsMemoryAndRejectsUnknownDrivers(t *testing.T) {
	sessions, counter, closeBackend, err := newSessionBackend(sessionSettings{driver: "memory"})
	if err != nil {
		t.Fatalf("memory backend: %v", err)
	}
	if sessions == nil || counter == nil {
		t.Fatal("memory backend returned a nil store or counter")
	}
	closeBackend()

	if _, _, _, err := newSessionBackend(sessionSettings{driver: "cookie-only"}); err == nil {
		t.Fatal("expected an error for an unsupported session driver")
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
