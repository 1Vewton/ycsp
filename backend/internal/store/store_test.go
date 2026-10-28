package store_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/1Vewton/ycsp/backend/internal/auth"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"gorm.io/gorm"
)

// 每个测试用临时目录里的 SQLite 文件，彼此不干扰。
func openMigratedDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := store.Open(store.Settings{
		Driver: store.DriverSQLite,
		DSN:    filepath.Join(t.TempDir(), "ycsp-test.db"),
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

// createTaskWithResult 造一条“用户 → 任务 → 结果”的完整链路。
func createTaskWithResult(t *testing.T, database *gorm.DB) (store.PredictionTask, store.StructureResult) {
	t.Helper()

	user := store.User{DisplayName: "小明"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	task := store.PredictionTask{UserID: &user.ID, Formula: "NaCl", Status: store.TaskStatusPending}
	if err := database.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	result := store.StructureResult{
		TaskID:             task.ID,
		Family:             store.FamilyIonic,
		Prototype:          "rock_salt",
		CoordinationNumber: 6,
		EdgeAAngstrom:      5.66,
		EdgeBAngstrom:      5.66,
		EdgeCAngstrom:      5.66,
		AngleAlphaDegree:   90,
		AngleBetaDegree:    90,
		AngleGammaDegree:   90,
		Atoms:              `[{"symbol":"Na","fractional_coordinates":[0,0,0]}]`,
		Rule:               `{"code":"rule_radius_ratio","params":{"ratio":0.564}}`,
		Caveats:            `[]`,
	}
	if err := database.Create(&result).Error; err != nil {
		t.Fatalf("create result: %v", err)
	}
	return task, result
}

func TestMigrateCreatesTheThreeTables(t *testing.T) {
	database := openMigratedDatabase(t)

	for _, table := range []string{"users", "prediction_tasks", "structure_results"} {
		if !database.Migrator().HasTable(table) {
			t.Errorf("missing table %q", table)
		}
	}
}

func TestMigrateCanRunTwice(t *testing.T) {
	database := openMigratedDatabase(t)

	if err := store.Migrate(database); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

func TestTaskLoadsWithItsUserAndResult(t *testing.T) {
	database := openMigratedDatabase(t)
	task, result := createTaskWithResult(t, database)

	startedAt := time.Now().UTC()
	durationMilliseconds := int64(12)
	updates := map[string]any{
		"status":      store.TaskStatusSucceeded,
		"started_at":  startedAt,
		"finished_at": startedAt,
		"duration_ms": durationMilliseconds,
	}
	if err := database.Model(&store.PredictionTask{}).
		Where("id = ?", task.ID).
		Updates(updates).Error; err != nil {
		t.Fatalf("update task: %v", err)
	}

	var loaded store.PredictionTask
	if err := database.Preload("User").First(&loaded, task.ID).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}

	var loadedResult store.StructureResult
	if err := database.Where("task_id = ?", task.ID).First(&loadedResult).Error; err != nil {
		t.Fatalf("load result: %v", err)
	}

	if loaded.Status != store.TaskStatusSucceeded {
		t.Errorf("status = %q, want %q", loaded.Status, store.TaskStatusSucceeded)
	}
	if loaded.StartedAt == nil || loaded.StartedAt.UnixMilli() != startedAt.UnixMilli() {
		t.Errorf("started_at = %v, want %v", loaded.StartedAt, startedAt)
	}
	if loaded.DurationMS == nil || *loaded.DurationMS != durationMilliseconds {
		t.Errorf("duration_ms = %v, want %d", loaded.DurationMS, durationMilliseconds)
	}
	if loaded.User == nil || loaded.User.DisplayName != "小明" {
		t.Errorf("user not preloaded correctly: %+v", loaded.User)
	}
	if loadedResult.Prototype != "rock_salt" {
		t.Errorf("result.prototype = %q, want %q", loadedResult.Prototype, "rock_salt")
	}
	if loadedResult.TaskID != result.TaskID {
		t.Errorf("result.task_id = %d, want %d", loadedResult.TaskID, result.TaskID)
	}
}

func TestOneResultPerTask(t *testing.T) {
	database := openMigratedDatabase(t)
	_, result := createTaskWithResult(t, database)

	duplicate := result
	duplicate.ID = 0
	if err := database.Create(&duplicate).Error; err == nil {
		t.Fatal("expected a unique constraint error for a second result on the same task")
	}
}

func TestDeletingTaskDeletesItsResult(t *testing.T) {
	database := openMigratedDatabase(t)
	task, _ := createTaskWithResult(t, database)

	if err := database.Delete(&store.PredictionTask{}, task.ID).Error; err != nil {
		t.Fatalf("delete task: %v", err)
	}

	var remaining int64
	if err := database.Model(&store.StructureResult{}).
		Where("task_id = ?", task.ID).
		Count(&remaining).Error; err != nil {
		t.Fatalf("count results: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("%d results left after deleting the task, want 0", remaining)
	}
}

func TestDeletingUserKeepsTheTaskWithoutOwner(t *testing.T) {
	database := openMigratedDatabase(t)
	task, _ := createTaskWithResult(t, database)

	if err := database.Delete(&store.User{}, *task.UserID).Error; err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var loaded store.PredictionTask
	if err := database.First(&loaded, task.ID).Error; err != nil {
		t.Fatalf("task should survive user deletion: %v", err)
	}
	if loaded.UserID != nil {
		t.Fatalf("user_id = %d, want NULL", *loaded.UserID)
	}
}

func TestSQLiteForeignKeysAreEnabled(t *testing.T) {
	database := openMigratedDatabase(t)

	var enabled int
	if err := database.Raw("PRAGMA foreign_keys").Scan(&enabled).Error; err != nil {
		t.Fatalf("read pragma: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1", enabled)
	}
}

func TestMigrateCreatesTheExpectedIndexes(t *testing.T) {
	database := openMigratedDatabase(t)

	expectations := []struct {
		model any
		index string
	}{
		{&store.PredictionTask{}, "idx_tasks_status_created"},
		{&store.PredictionTask{}, "idx_prediction_tasks_user_id"},
		{&store.StructureResult{}, "idx_structure_results_task_id"},
	}

	for _, expectation := range expectations {
		if !database.Migrator().HasIndex(expectation.model, expectation.index) {
			t.Errorf("missing index %q", expectation.index)
		}
	}
}

func TestMigrateReportsFailureOnAClosedDatabase(t *testing.T) {
	database, err := store.Open(store.Settings{
		Driver: store.DriverSQLite,
		DSN:    filepath.Join(t.TempDir(), "closed.db"),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	sqlDatabase, err := database.DB()
	if err != nil {
		t.Fatalf("read connection pool: %v", err)
	}
	if err := sqlDatabase.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	if err := store.Migrate(database); err == nil {
		t.Fatal("expected an error when migrating a closed database")
	}
}

func TestUserStoresOnlyThePasswordHash(t *testing.T) {
	database := openMigratedDatabase(t)

	const plainPassword = "ycsp-student-passphrase"
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := store.User{DisplayName: "小明", PasswordHash: hashedPassword}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	var loaded store.User
	if err := database.First(&loaded, user.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if loaded.PasswordHash != hashedPassword {
		t.Error("password hash changed after a database round trip")
	}
	if !auth.VerifyPassword(loaded.PasswordHash, plainPassword) {
		t.Error("stored hash no longer verifies the password")
	}

	var plaintextRows int64
	if err := database.Model(&store.User{}).
		Where("password_hash = ?", plainPassword).
		Count(&plaintextRows).Error; err != nil {
		t.Fatalf("count plaintext rows: %v", err)
	}
	if plaintextRows != 0 {
		t.Error("the plaintext password was stored in the database")
	}
}

func TestMigrateAddsPasswordHashToAnExistingUsersTable(t *testing.T) {
	// 这个测试要从“没有 password_hash 的旧库”开始，所以不能先用 Migrate 建表。
	database, err := store.Open(store.Settings{
		Driver: store.DriverSQLite,
		DSN:    filepath.Join(t.TempDir(), "legacy.db"),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		sqlDatabase, err := database.DB()
		if err == nil {
			_ = sqlDatabase.Close()
		}
	})

	// 模拟账号体系之前的 users 表：没有 password_hash 列，而且已经有一条记录。
	if err := database.Exec(`
		CREATE TABLE users (
			id integer primary key,
			display_name text not null,
			created_at datetime
		)`).Error; err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if err := database.Exec(
		`INSERT INTO users (display_name, created_at) VALUES (?, CURRENT_TIMESTAMP)`, "小明",
	).Error; err != nil {
		t.Fatalf("insert legacy user: %v", err)
	}

	if err := store.Migrate(database); err != nil {
		t.Fatalf("migrate legacy database: %v", err)
	}

	var loaded store.User
	if err := database.First(&loaded).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if loaded.PasswordHash != "" {
		t.Errorf("password_hash = %q, want empty for legacy accounts", loaded.PasswordHash)
	}
	if auth.VerifyPassword(loaded.PasswordHash, "anything") {
		t.Error("a legacy account without a password must not pass verification")
	}
}

func TestOpenRejectsUnsupportedDriver(t *testing.T) {
	if _, err := store.Open(store.Settings{Driver: "oracle", DSN: "xe"}); err == nil {
		t.Fatal("expected an error for an unsupported driver")
	}
}

func TestOpenRejectsEmptyDSN(t *testing.T) {
	if _, err := store.Open(store.Settings{Driver: store.DriverSQLite, DSN: "  "}); err == nil {
		t.Fatal("expected an error for an empty DSN")
	}
}
