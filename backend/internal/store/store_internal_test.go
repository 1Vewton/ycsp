package store

import "testing"

func TestWithSQLiteForeignKeysAddsThePragma(t *testing.T) {
	testCases := []struct {
		name string
		dsn  string
		want string
	}{
		{"plain file", "ycsp.db", "ycsp.db?_pragma=foreign_keys(1)"},
		{"file uri", "file:ycsp.db", "file:ycsp.db?_pragma=foreign_keys(1)"},
		{"existing query", "file:ycsp.db?mode=memory", "file:ycsp.db?mode=memory&_pragma=foreign_keys(1)"},
		{"already configured", "ycsp.db?_pragma=busy_timeout(5000)", "ycsp.db?_pragma=busy_timeout(5000)"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := withSQLiteForeignKeys(testCase.dsn); got != testCase.want {
				t.Fatalf("withSQLiteForeignKeys(%q) = %q, want %q", testCase.dsn, got, testCase.want)
			}
		})
	}
}

func TestNewDialectorSupportsEveryConfiguredDriver(t *testing.T) {
	// 只检查“能构造出 dialector”，不真的连库：CI 里没有 postgres / mysql 服务。
	for _, driver := range []string{DriverSQLite, DriverPostgres, DriverMySQL} {
		t.Run(driver, func(t *testing.T) {
			dialector, err := newDialector(Settings{Driver: driver, DSN: "example-dsn"})
			if err != nil {
				t.Fatalf("newDialector(%q): %v", driver, err)
			}
			if dialector == nil {
				t.Fatalf("newDialector(%q) returned nil", driver)
			}
		})
	}
}
