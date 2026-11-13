// Package testutil 给各包的测试提供一次性数据库。
//
// 默认用临时 SQLite（本地零依赖）；设置 YCSP_TEST_DB_DRIVER=postgres 与
// YCSP_TEST_DB_DSN 后改用 Postgres，每个测试用独立 schema 隔离。
// 它不是生产代码，只被 *_test.go 引用。
package testutil

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/store"
)

// postgresSchemaCounter 让同一进程里并发创建的 schema 名称互不重复。
var postgresSchemaCounter uint64

// OpenDatabase 建一个迁移好三张表的临时数据库，测试结束自动关闭。
// 默认 SQLite；设置 YCSP_TEST_DB_DRIVER=postgres 时用 Postgres 独立 schema。
func OpenDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	driver := strings.TrimSpace(os.Getenv("YCSP_TEST_DB_DRIVER"))
	if driver == "" {
		driver = store.DriverSQLite
	}

	switch driver {
	case store.DriverSQLite:
		dsn := strings.TrimSpace(os.Getenv("YCSP_TEST_DB_DSN"))
		if dsn == "" {
			dsn = filepath.Join(t.TempDir(), "test.db")
		}
		return openMigrated(t, store.Settings{Driver: driver, DSN: dsn})
	case store.DriverPostgres:
		return openPostgresSchema(t, os.Getenv("YCSP_TEST_DB_DSN"))
	default:
		t.Fatalf(
			"testutil: unsupported YCSP_TEST_DB_DRIVER %q, choose %s or %s",
			driver, store.DriverSQLite, store.DriverPostgres,
		)
		return nil
	}
}

// openMigrated 打开数据库并建表；SQLite 与 Postgres 共用这段收尾逻辑。
func openMigrated(t *testing.T, settings store.Settings) *gorm.DB {
	t.Helper()

	database, err := store.Open(settings)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := store.Migrate(database); err != nil {
		closeDatabase(database)
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { closeDatabase(database) })
	return database
}

// openPostgresSchema 在测试库里建一个独立 schema，迁移好后返回只使用该 schema 的连接。
// 这样同一个 Postgres 服务可以并行跑多个包/测试，不会互相看到对方的表。
func openPostgresSchema(t *testing.T, baseDSN string) *gorm.DB {
	t.Helper()

	trimmedDSN := strings.TrimSpace(baseDSN)
	if trimmedDSN == "" {
		t.Fatal("testutil: YCSP_TEST_DB_DSN is required when YCSP_TEST_DB_DRIVER=postgres")
	}

	admin, err := store.Open(store.Settings{
		Driver: store.DriverPostgres,
		DSN:    trimmedDSN,
	})
	if err != nil {
		t.Fatalf("open postgres admin connection: %v", err)
	}

	schema := fmt.Sprintf(
		"ycsp_test_%d_%d_%d",
		os.Getpid(),
		time.Now().UnixNano(),
		atomic.AddUint64(&postgresSchemaCounter, 1),
	)
	quotedSchema := quoteIdentifier(schema)
	if err := admin.Exec("CREATE SCHEMA " + quotedSchema).Error; err != nil {
		closeDatabase(admin)
		t.Fatalf("create test schema %s: %v", schema, err)
	}

	// 先登记 admin 的清理，再登记测试连接的清理；t.Cleanup 后进先出，
	// 因此测试连接会先关闭，然后才删除 schema。
	t.Cleanup(func() {
		if err := admin.Exec("DROP SCHEMA " + quotedSchema + " CASCADE").Error; err != nil {
			t.Logf("drop test schema %s: %v", schema, err)
		}
		closeDatabase(admin)
	})

	database, err := store.Open(store.Settings{
		Driver: store.DriverPostgres,
		DSN:    withSearchPath(trimmedDSN, schema),
	})
	if err != nil {
		t.Fatalf("open postgres test schema %s: %v", schema, err)
	}
	if err := store.Migrate(database); err != nil {
		closeDatabase(database)
		t.Fatalf("migrate postgres test schema %s: %v", schema, err)
	}

	t.Cleanup(func() { closeDatabase(database) })
	return database
}

// withSearchPath 让 Postgres 的每条连接都默认使用测试 schema。
// pgx 会把未识别的连接串参数当作运行时参数，因此 search_path 可以直接放在 DSN 里。
func withSearchPath(dsn string, schema string) string {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err == nil {
			query := parsed.Query()
			query.Set("search_path", schema)
			parsed.RawQuery = query.Encode()
			return parsed.String()
		}
	}
	return dsn + " search_path=" + schema
}

// quoteIdentifier 安全地引用 Postgres 标识符；schema 名由本文件生成，仍按规则转义。
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func closeDatabase(database *gorm.DB) {
	sqlDatabase, err := database.DB()
	if err == nil {
		_ = sqlDatabase.Close()
	}
}
