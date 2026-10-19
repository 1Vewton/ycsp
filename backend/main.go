// 数据管理层（Go + Gin）入口：读取配置并启动服务。
package main

import (
	"fmt"
	"log"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/env"
)

// .env 未提供对应项时使用的默认值。
const (
	defaultListenHost     = "127.0.0.1"
	defaultListenPort     = 8000
	defaultComputeBaseURL = "http://127.0.0.1:8001"

	envFilePath = ".env"
)

// appSettings 是数据管理层的运行配置。
type appSettings struct {
	listenHost     string
	listenPort     int
	computeBaseURL string
}

func main() {
	env.Load(envFilePath)

	settings, err := readSettings()
	if err != nil {
		log.Fatal(err)
	}

	router := api.NewRouter(settings.computeBaseURL)

	address := fmt.Sprintf("%s:%d", settings.listenHost, settings.listenPort)
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

	return appSettings{
		listenHost:     env.Get("YCSP_BACKEND_HOST", defaultListenHost),
		listenPort:     listenPort,
		computeBaseURL: env.Get("YCSP_COMPUTE_BASE_URL", defaultComputeBaseURL),
	}, nil
}
