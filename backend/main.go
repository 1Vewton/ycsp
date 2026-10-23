// 数据管理层（Go + Gin）入口：读取配置并启动服务。
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/env"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
)

// .env 未提供对应项时使用的默认值。
const (
	defaultListenHost     = "127.0.0.1"
	defaultListenPort     = 8000
	defaultComputeBaseURL = "http://127.0.0.1:8001"
	// 默认用 SQLite 文件：零安装就能跑；正式部署改 YCSP_DB_DRIVER / YCSP_DB_DSN 即可。
	defaultDatabaseDriver = store.DriverSQLite
	defaultDatabaseDSN    = "ycsp.db"

	envFilePath = ".env"
)

// appSettings 是数据管理层的运行配置。
type appSettings struct {
	listenHost     string
	listenPort     int
	computeBaseURL string
	database       store.Settings
}

func main() {
	env.Load(envFilePath)

	settings, err := readSettings()
	if err != nil {
		log.Fatal(err)
	}

	database, err := store.Open(settings.database)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Migrate(database); err != nil {
		log.Fatal(err)
	}
	sqlDatabase, err := database.DB()
	if err != nil {
		log.Fatalf("db: read connection pool: %v", err)
	}
	defer sqlDatabase.Close()

	taskService := tasks.New(database, compute.NewClient(settings.computeBaseURL))
	workerContext, stopWorkers := context.WithCancel(context.Background())
	taskService.Start(workerContext)
	// defer 后进先出：先停 worker，再由更早注册的 defer 关闭数据库连接。
	defer func() {
		stopWorkers()
		taskService.Stop()
	}()

	router := api.NewRouter(api.Options{
		ComputeBaseURL: settings.computeBaseURL,
		Tasks:          taskService,
	})

	address := fmt.Sprintf("%s:%d", settings.listenHost, settings.listenPort)
	// 只打印驱动名，不打印 DSN：连接串里可能有密码。
	log.Printf("%s database driver: %s", api.ServiceName, settings.database.Driver)
	log.Printf("%s listening on http://%s", api.ServiceName, address)
	if err := router.Run(address); err != nil {
		log.Fatalf("%s failed to start: %v", api.ServiceName, err)
	}
}

// readSettings 读取运行配置，缺项时用默认值。
func readSettings() (appSettings, error) {
	listenPort, err := env.GetInt("YCSP_BACKEND_PORT", defaultListenPort)
	if err != nil {
		return appSettings{}, err
	}

	// SQLite 有默认文件名；其他引擎的 DSN 含地址与账号，不能瞎猜，必须显式给出。
	databaseDriver := env.Get("YCSP_DB_DRIVER", defaultDatabaseDriver)
	databaseDSN := env.Get("YCSP_DB_DSN", "")
	if databaseDSN == "" && databaseDriver == store.DriverSQLite {
		databaseDSN = defaultDatabaseDSN
	}

	return appSettings{
		listenHost:     env.Get("YCSP_BACKEND_HOST", defaultListenHost),
		listenPort:     listenPort,
		computeBaseURL: env.Get("YCSP_COMPUTE_BASE_URL", defaultComputeBaseURL),
		database: store.Settings{
			Driver: databaseDriver,
			DSN:    databaseDSN,
		},
	}, nil
}
