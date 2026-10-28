// Package env 统一读取环境变量，未设置时使用默认值。
package env

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Load 读取 .env 文件；读不到时只提示一句，随后仍用系统环境变量与默认值。
func Load(path string) {
	if err := godotenv.Load(path); err != nil {
		log.Printf("env: %s not loaded (%v); using system environment and defaults", path, err)
	}
}

// Get 读取字符串环境变量；未设置或为空时返回 fallback。
func Get(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// GetInt 读取整数环境变量；未设置或为空时返回 fallback。
func GetInt(key string, fallback int) (int, error) {
	raw := Get(key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("env: %s must be an integer, got %q", key, raw)
	}
	return value, nil
}

// GetBool 读取布尔环境变量；未设置或为空时返回 fallback。
// 只认 true/false/1/0（大小写不敏感）：拼错的 "ture" 直接报错，而不是悄悄当 false。
func GetBool(key string, fallback bool) (bool, error) {
	raw := Get(key, strconv.FormatBool(fallback))
	switch strings.ToLower(raw) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("env: %s must be true/false, got %q", key, raw)
	}
}
