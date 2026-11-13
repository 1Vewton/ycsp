// Package store 负责数据库连接、GORM 模型与自动迁移。
//
// 表结构的设计（字段、约束、状态机）见 backend/README.md《数据表》；
// 这里只落地：模型、连接、迁移。
package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 支持的数据库引擎，取值来自 YCSP_DB_DRIVER。
const (
	DriverSQLite   = "sqlite"
	DriverPostgres = "postgres"
	DriverMySQL    = "mysql"
)

// sqliteForeignKeysPragma 打开 SQLite 的外键检查：它默认是关闭的，
// 不打开的话 ON DELETE CASCADE / SET NULL 都不会生效。
const sqliteForeignKeysPragma = "_pragma=foreign_keys(1)"

// SQLite 同一时刻只允许一个写入者，连接开多了会频繁报 "database is locked"。
// 它只用于本地开发和测试，所以这里固定单连接；生产用 Postgres / MySQL 的连接池。
const (
	sqliteMaxOpenConnections = 1
	sqliteMaxIdleConnections = 1
)

// Postgres / MySQL 的默认连接池。教学场景并发不高，20 个连接足够；
// 生产可以按数据库的 max_connections 和实际压测结果调整。
const (
	defaultServerMaxOpenConnections    = 20
	defaultServerMaxIdleConnections    = 10
	defaultServerConnectionMaxLifetime = 30 * time.Minute
)

// databasePingTimeout 是启动时检查数据库是否可用的上限：
// 连不上时尽快失败，避免服务看起来启动成功、第一次请求才发现问题。
const databasePingTimeout = 5 * time.Second

// Settings 是数据库连接配置。
type Settings struct {
	Driver string // sqlite / postgres / mysql
	DSN    string // 各驱动自己的连接串；sqlite 是数据库文件路径

	// 连接池参数；0 表示使用驱动默认值。SQLite 会忽略这些值并固定为单连接。
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// poolSettings 是最终应用到 database/sql 的连接池参数。
type poolSettings struct {
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
}

// connectionPoolSettings 按驱动决定连接池参数：
// SQLite 固定单连接；Postgres / MySQL 使用配置值或服务端默认值。
func connectionPoolSettings(settings Settings) poolSettings {
	if settings.Driver == DriverSQLite {
		return poolSettings{
			maxOpen: sqliteMaxOpenConnections,
			maxIdle: sqliteMaxIdleConnections,
		}
	}

	result := poolSettings{
		maxOpen:     defaultServerMaxOpenConnections,
		maxIdle:     defaultServerMaxIdleConnections,
		maxLifetime: defaultServerConnectionMaxLifetime,
	}
	if settings.MaxOpenConns > 0 {
		result.maxOpen = settings.MaxOpenConns
	}
	if settings.MaxIdleConns > 0 {
		result.maxIdle = settings.MaxIdleConns
	}
	if settings.ConnMaxLifetime > 0 {
		result.maxLifetime = settings.ConnMaxLifetime
	}
	if result.maxIdle > result.maxOpen {
		result.maxIdle = result.maxOpen
	}
	return result
}

// Open 按配置建立数据库连接；驱动名或连接串有问题时立刻返回错误。
func Open(settings Settings) (*gorm.DB, error) {
	dialector, err := newDialector(settings)
	if err != nil {
		return nil, err
	}

	database, err := gorm.Open(dialector, &gorm.Config{
		// 只记录慢查询与错误，避免每个请求都刷一行 SQL。
		Logger: logger.Default.LogMode(logger.Warn),
		// 把驱动的唯一键冲突翻译成 gorm.ErrDuplicatedKey：注册重名时才能准确报错。
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", settings.Driver, err)
	}

	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("db: read connection pool: %w", err)
	}

	pool := connectionPoolSettings(settings)
	sqlDatabase.SetMaxOpenConns(pool.maxOpen)
	sqlDatabase.SetMaxIdleConns(pool.maxIdle)
	sqlDatabase.SetConnMaxLifetime(pool.maxLifetime)

	pingContext, cancelPing := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancelPing()
	if err := sqlDatabase.PingContext(pingContext); err != nil {
		_ = sqlDatabase.Close()
		return nil, fmt.Errorf("db: ping %s: %w", settings.Driver, err)
	}
	return database, nil
}

// Migrate 建表或补齐缺失的列；服务启动时调用，可重复执行。
func Migrate(database *gorm.DB) error {
	if err := database.AutoMigrate(&User{}, &PredictionTask{}, &StructureResult{}); err != nil {
		return fmt.Errorf("db: migrate: %w", err)
	}
	return nil
}

// EnsureSchema 检查三张表是否都已存在。
// 生产默认关闭自动迁移，启动前用它挡住“忘了跑 -migrate-only”的部署。
func EnsureSchema(database *gorm.DB) error {
	for _, table := range []string{"users", "prediction_tasks", "structure_results"} {
		if !database.Migrator().HasTable(table) {
			return fmt.Errorf(
				"db: table %q is missing; run `go run . -migrate-only` first", table,
			)
		}
	}
	return nil
}

// newDialector 按配置选择驱动。函数未导出，测试用 newDialector 覆盖三种引擎。
func newDialector(settings Settings) (gorm.Dialector, error) {
	if strings.TrimSpace(settings.DSN) == "" {
		return nil, fmt.Errorf("db: YCSP_DB_DSN is empty")
	}

	switch settings.Driver {
	case DriverSQLite:
		return sqlite.Open(withSQLiteForeignKeys(settings.DSN)), nil
	case DriverPostgres:
		return postgres.Open(settings.DSN), nil
	case DriverMySQL:
		return mysql.Open(settings.DSN), nil
	default:
		return nil, fmt.Errorf(
			"db: unsupported driver %q, choose one of %s / %s / %s",
			settings.Driver, DriverSQLite, DriverPostgres, DriverMySQL,
		)
	}
}

// withSQLiteForeignKeys 给 SQLite 连接串补上外键开关。
// 使用者自己写了 _pragma 就不动，避免覆盖他有意设置的参数。
func withSQLiteForeignKeys(dsn string) string {
	if strings.Contains(dsn, "_pragma") {
		return dsn
	}

	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + sqliteForeignKeysPragma
}
