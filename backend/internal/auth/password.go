// Package auth 处理账号安全：密码哈希与校验。
//
// 只存 bcrypt 哈希，从不保存明文。bcrypt 每次都会生成随机盐并写进哈希串本身，
// 所以数据库只要一列 password_hash：盐与哈希始终配套，也不会出现“盐换错”的问题。
package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordEmpty 表示没填密码。
	ErrPasswordEmpty = errors.New("auth: password is empty")
	// ErrPasswordTooLong 表示密码超过 bcrypt 能处理的长度。
	ErrPasswordTooLong = fmt.Errorf("auth: password must be at most %d bytes", MaxPasswordBytes)
)

// dummyHash 让“账号不存在”也走一次代价相同的 bcrypt 校验，避免用响应时间探测账号。
var dummyHash = generateDummyHash()

func generateDummyHash() []byte {
	hashed, err := bcrypt.GenerateFromPassword([]byte("ycsp-dummy-password-for-timing"), bcrypt.DefaultCost)
	if err != nil {
		// 常量输入不会失败；真失败说明环境异常，宁可启动失败也不要留下时序侧信道。
		panic(fmt.Sprintf("auth: build dummy hash: %v", err))
	}
	return hashed
}

// VerifyAbsentAccount 在“账号不存在”时也消耗一次校验时间。
func VerifyAbsentAccount(plainPassword string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plainPassword))
}

// HashPassword 把明文密码变成 bcrypt 哈希；每次调用都会用新的随机盐。
func HashPassword(plainPassword string) (string, error) {
	if plainPassword == "" {
		return "", ErrPasswordEmpty
	}
	if len(plainPassword) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}

	// 默认代价因子 10：单次约几十毫秒，登录时感觉不到，但足以让离线爆破变慢。
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword 校验明文与哈希是否匹配。
// 哈希为空（历史上没设过密码的账号）或格式不对时一律返回 false，
// 不区分具体原因，避免向外泄露“这个账号存在/没设密码”之类信息。
func VerifyPassword(hashedPassword string, plainPassword string) bool {
	if hashedPassword == "" || plainPassword == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
