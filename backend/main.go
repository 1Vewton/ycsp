// 数据管理层（Go + Gin）入口：对外提供统一接口，并把计算请求转发给数据处理层。
package main

import (
	"log"
	"os"

	"github.com/1Vewton/ycsp/backend/internal/api"
)

const (
	defaultListenPort     = "8000"
	defaultComputeBaseURL = "http://127.0.0.1:8001"
)

func main() {
	listenPort := getEnvOrDefault("YCSP_BACKEND_PORT", defaultListenPort)
	computeBaseURL := getEnvOrDefault("YCSP_COMPUTE_BASE_URL", defaultComputeBaseURL)

	router := api.NewRouter(computeBaseURL)

	log.Printf("%s 已启动：http://127.0.0.1:%s", api.ServiceName, listenPort)
	if err := router.Run("127.0.0.1:" + listenPort); err != nil {
		log.Fatalf("%s 启动失败：%v", api.ServiceName, err)
	}
}

// getEnvOrDefault 读环境变量；未设置或为空时返回 fallback。
func getEnvOrDefault(key string, fallback string) string {
	if value, isSet := os.LookupEnv(key); isSet && value != "" {
		return value
	}
	return fallback
}
