// 数据管理层（Go + Gin）入口：读取配置并启动服务。
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/env"
	"github.com/1Vewton/ycsp/backend/internal/ratelimit"
	"github.com/1Vewton/ycsp/backend/internal/session"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
	"github.com/1Vewton/ycsp/backend/internal/users"
)

// .env 未提供对应项时使用的默认值。
const (
	defaultListenHost     = "127.0.0.1"
	defaultListenPort     = 8000
	defaultComputeBaseURL = "http://127.0.0.1:8001"
	// 默认用 SQLite 文件：零安装就能跑；正式部署改 YCSP_DB_DRIVER / YCSP_DB_DSN 即可。
	defaultDatabaseDriver = store.DriverSQLite
	defaultDatabaseDSN    = "ycsp.db"
	// 会话默认存 Redis；本地没有 Redis 时可设 YCSP_SESSION_DRIVER=memory（仅单机开发）。
	defaultSessionDriver          = session.DriverRedis
	defaultRedisURL               = "redis://127.0.0.1:6379/0"
	defaultSessionLifetimeSeconds = 7 * 24 * 60 * 60 // 7 天
	defaultCookieSecure           = false
	// 登录失败限流：同一昵称 15 分钟内失败 5 次、同一地址失败 20 次即锁定 15 分钟。
	defaultLoginMaxAttemptsPerAccount = 5
	defaultLoginMaxAttemptsPerAddress = 20
	defaultLoginBlockSeconds          = 15 * 60

	redisPingTimeout = 3 * time.Second

	envFilePath = ".env"
)

// appSettings 是数据管理层的运行配置。
type appSettings struct {
	listenHost     string
	listenPort     int
	computeBaseURL string
	database       store.Settings
	sessions       sessionSettings
	loginLimit     loginLimitSettings
}

// sessionSettings 是登录会话的配置。
type sessionSettings struct {
	driver       string
	redisURL     string
	lifetime     time.Duration
	secureCookie bool
}

// loginLimitSettings 是登录失败限流的阈值。
type loginLimitSettings struct {
	maxAttemptsPerAccount int
	maxAttemptsPerAddress int
	window                time.Duration
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

	sessionStore, loginCounter, closeSessionBackend, err := newSessionBackend(settings.sessions)
	if err != nil {
		log.Fatal(err)
	}
	defer closeSessionBackend()

	loginGuard, err := ratelimit.NewGuard(loginCounter, ratelimit.Policy{
		MaxAttemptsPerAccount: settings.loginLimit.maxAttemptsPerAccount,
		MaxAttemptsPerAddress: settings.loginLimit.maxAttemptsPerAddress,
		Window:                settings.loginLimit.window,
	})
	if err != nil {
		log.Fatal(err)
	}

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
		Auth: api.AuthOptions{
			Users:           users.New(database),
			Sessions:        sessionStore,
			SessionLifetime: settings.sessions.lifetime,
			SecureCookies:   settings.sessions.secureCookie,
			LoginGuard:      loginGuard,
		},
	})

	address := fmt.Sprintf("%s:%d", settings.listenHost, settings.listenPort)
	// 只打印驱动名，不打印 DSN：连接串里可能有密码。
	log.Printf("%s database driver: %s", api.ServiceName, settings.database.Driver)
	log.Printf("%s session store: %s", api.ServiceName, settings.sessions.driver)
	log.Printf("%s listening on http://%s", api.ServiceName, address)
	if err := router.Run(address); err != nil {
		log.Fatalf("%s failed to start: %v", api.ServiceName, err)
	}
}

// newSessionBackend 按配置创建会话存储与登录失败计数器（两者共用同一个后端）。
// 返回的 close 在进程退出时释放连接。Redis 连不上时直接失败：
// 登录功能半死不活比启动失败更难排查。
func newSessionBackend(
	settings sessionSettings,
) (session.Store, ratelimit.Counter, func(), error) {
	switch settings.driver {
	case session.DriverRedis:
		options, err := redis.ParseURL(settings.redisURL)
		if err != nil {
			return nil, nil, nil, fmt.Errorf(
				"session: invalid YCSP_REDIS_URL %q: %w", settings.redisURL, err,
			)
		}
		client := redis.NewClient(options)

		pingContext, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
		defer cancel()
		if err := client.Ping(pingContext).Err(); err != nil {
			_ = client.Close()
			return nil, nil, nil, fmt.Errorf(
				"session: redis unreachable at %s: %w (本地没有 Redis 时可设 YCSP_SESSION_DRIVER=memory)",
				settings.redisURL, err,
			)
		}
		return session.NewRedisStore(client),
			ratelimit.NewRedisCounter(client),
			func() { _ = client.Close() },
			nil

	case session.DriverMemory:
		log.Printf("warning: sessions and login counters live in process memory; restarting signs everyone out")
		return session.NewMemoryStore(), ratelimit.NewMemoryCounter(), func() {}, nil

	default:
		return nil, nil, nil, fmt.Errorf(
			"session: unsupported YCSP_SESSION_DRIVER %q, choose %s or %s",
			settings.driver, session.DriverRedis, session.DriverMemory,
		)
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

	sessionLifetimeSeconds, err := env.GetInt(
		"YCSP_SESSION_LIFETIME_SECONDS", defaultSessionLifetimeSeconds,
	)
	if err != nil {
		return appSettings{}, err
	}
	cookieSecure, err := env.GetBool("YCSP_COOKIE_SECURE", defaultCookieSecure)
	if err != nil {
		return appSettings{}, err
	}

	loginMaxAttempts, err := env.GetInt(
		"YCSP_LOGIN_MAX_ATTEMPTS", defaultLoginMaxAttemptsPerAccount,
	)
	if err != nil {
		return appSettings{}, err
	}
	loginAddressAttempts, err := env.GetInt(
		"YCSP_LOGIN_IP_MAX_ATTEMPTS", defaultLoginMaxAttemptsPerAddress,
	)
	if err != nil {
		return appSettings{}, err
	}
	loginBlockSeconds, err := env.GetInt("YCSP_LOGIN_BLOCK_SECONDS", defaultLoginBlockSeconds)
	if err != nil {
		return appSettings{}, err
	}

	return appSettings{
		listenHost:     env.Get("YCSP_BACKEND_HOST", defaultListenHost),
		listenPort:     listenPort,
		computeBaseURL: env.Get("YCSP_COMPUTE_BASE_URL", defaultComputeBaseURL),
		database: store.Settings{
			Driver: databaseDriver,
			DSN:    databaseDSN,
		},
		sessions: sessionSettings{
			driver:       env.Get("YCSP_SESSION_DRIVER", defaultSessionDriver),
			redisURL:     env.Get("YCSP_REDIS_URL", defaultRedisURL),
			lifetime:     time.Duration(sessionLifetimeSeconds) * time.Second,
			secureCookie: cookieSecure,
		},
		loginLimit: loginLimitSettings{
			maxAttemptsPerAccount: loginMaxAttempts,
			maxAttemptsPerAddress: loginAddressAttempts,
			window:                time.Duration(loginBlockSeconds) * time.Second,
		},
	}, nil
}
