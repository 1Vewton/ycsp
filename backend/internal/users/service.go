// Package users 处理账号注册与登录校验。
package users

import (
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/auth"
	"github.com/1Vewton/ycsp/backend/internal/store"
)

const (
	// MinDisplayNameLength / MaxDisplayNameLength 是昵称长度（字符数）限制。
	// 昵称同时用来登录，所以要求唯一。
	MinDisplayNameLength = 2
	MaxDisplayNameLength = 32
)

var (
	// ErrNameTaken 表示昵称已被占用。
	ErrNameTaken = errors.New("users: display name already taken")
	// ErrInvalidCredentials 表示昵称或密码不对；不区分两者，避免泄露账号是否存在。
	ErrInvalidCredentials = errors.New("users: invalid credentials")
	// ErrUserNotFound 表示用户编号不存在。
	ErrUserNotFound = errors.New("users: not found")
)

// InvalidNameError 表示昵称长度不符合要求。
type InvalidNameError struct {
	MinLength int
	MaxLength int
}

func (err *InvalidNameError) Error() string {
	return "users: display name length out of range"
}

// WeakPasswordError 列出所有没被满足的密码规则。
type WeakPasswordError struct {
	Rules []auth.PasswordRule
}

func (err *WeakPasswordError) Error() string {
	return "users: password does not meet the policy"
}

// Service 负责账号的读写。
type Service struct {
	database *gorm.DB
}

// New 创建账号服务。
func New(database *gorm.DB) *Service {
	return &Service{database: database}
}

// Register 创建账号：昵称唯一，密码只存 bcrypt 哈希。
func (service *Service) Register(displayName string, password string) (*store.User, error) {
	name := strings.TrimSpace(displayName)
	length := utf8.RuneCountInString(name)
	if length < MinDisplayNameLength || length > MaxDisplayNameLength {
		return nil, &InvalidNameError{
			MinLength: MinDisplayNameLength,
			MaxLength: MaxDisplayNameLength,
		}
	}
	if unmetRules := auth.CheckPasswordPolicy(password); len(unmetRules) > 0 {
		return nil, &WeakPasswordError{Rules: unmetRules}
	}

	var existing int64
	if err := service.database.Model(&store.User{}).
		Where("display_name = ?", name).
		Count(&existing).Error; err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, ErrNameTaken
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := store.User{DisplayName: name, PasswordHash: hashedPassword}
	if err := service.database.Create(&user).Error; err != nil {
		// 两个人同时抢同一个昵称时，靠唯一索引兜底。
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrNameTaken
		}
		return nil, err
	}
	return &user, nil
}

// Authenticate 校验昵称与密码；任何不匹配都返回 ErrInvalidCredentials。
func (service *Service) Authenticate(displayName string, password string) (*store.User, error) {
	var user store.User
	err := service.database.
		Where("display_name = ?", strings.TrimSpace(displayName)).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 账号不存在也做一次同样代价的校验，避免用响应时间探测账号是否存在。
			auth.VerifyAbsentAccount(password)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 历史上没设密码的账号哈希为空，VerifyPassword 会直接返回 false。
	if !auth.VerifyPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return &user, nil
}

// ByID 按编号取用户；不存在时返回 ErrUserNotFound。
func (service *Service) ByID(userID uint) (*store.User, error) {
	var user store.User
	if err := service.database.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
