// Package store 负责数据库连接、GORM 模型与自动迁移。
//
// 表结构的设计（字段、约束、状态机）见 backend/README.md《数据表》；
// 这里只落地：模型、连接、迁移。
package store

import (
	"fmt"
	"strings"

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

// sqliteMaxOpenConnections 限制 SQLite 的连接数：它同一时刻只允许一个写入者，
// 连接开多了会频繁报 "database is locked"。
const sqliteMaxOpenConnections = 1

// Settings 是数据库连接配置。
type Settings struct {
	Driver string // sqlite / postgres / mysql
	DSN    string // 各驱动自己的连接串；sqlite 是数据库文件路径
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

	if settings.Driver == DriverSQLite {
		sqlDatabase, err := database.DB()
		if err != nil {
			return nil, fmt.Errorf("db: read connection pool: %w", err)
		}
		sqlDatabase.SetMaxOpenConns(sqliteMaxOpenConnections)
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
