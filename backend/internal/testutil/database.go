// Package testutil 给各包的测试提供一个用完即弃的 SQLite 库。
//
// 它不是生产代码，只被 *_test.go 引用；单独成包是为了避免每个测试文件重复写建库逻辑。
package testutil

import (
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/store"
)

// OpenDatabase 建一个临时 SQLite 库并迁移好三张表，测试结束自动关闭。
func OpenDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := store.Open(store.Settings{
		Driver: store.DriverSQLite,
		DSN:    filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := store.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() {
		sqlDatabase, err := database.DB()
		if err == nil {
			_ = sqlDatabase.Close()
		}
	})
	return database
}
