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

	// 安全问题与答案的长度限制：问题按字符数算，答案受 bcrypt 的 72 字节上限约束。
	MinSecurityQuestionLength = 2
	MaxSecurityQuestionLength = 80
	MinSecurityAnswerLength   = 2
	MaxSecurityAnswerBytes    = auth.MaxPasswordBytes
)

var (
	// ErrNameTaken 表示昵称已被占用。
	ErrNameTaken = errors.New("users: display name already taken")
	// ErrInvalidCredentials 表示昵称或密码不对；不区分两者，避免泄露账号是否存在。
	ErrInvalidCredentials = errors.New("users: invalid credentials")
	// ErrUserNotFound 表示用户编号不存在。
	ErrUserNotFound = errors.New("users: not found")
	// ErrSecuritySetupIncomplete 表示安全问题与答案只填了一个。
	ErrSecuritySetupIncomplete = errors.New("users: security question and answer must be set together")
	// ErrCurrentPasswordIncorrect 表示修改密码或安全问题时的当前密码不对。
	ErrCurrentPasswordIncorrect = errors.New("users: current password is incorrect")
	// ErrPasswordUnchanged 表示新密码与当前密码相同。
	ErrPasswordUnchanged = errors.New("users: new password matches the current one")
	// ErrPasswordResetUnavailable 表示账号不存在或没有设置安全问题。
	ErrPasswordResetUnavailable = errors.New("users: password reset is not available")
	// ErrPasswordResetFailed 表示昵称、安全问题答案不匹配；统一返回，避免泄露账号是否存在。
	ErrPasswordResetFailed = errors.New("users: password reset failed")
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

// InvalidSecurityQuestionError 表示安全问题长度不符合要求。
type InvalidSecurityQuestionError struct {
	MinLength int
	MaxLength int
}

func (err *InvalidSecurityQuestionError) Error() string {
	return "users: security question length out of range"
}

// InvalidSecurityAnswerError 表示安全问题答案长度不符合要求。
type InvalidSecurityAnswerError struct {
	MinLength int
	MaxBytes  int
}

func (err *InvalidSecurityAnswerError) Error() string {
	return "users: security answer length out of range"
}

// Service 负责账号的读写。
type Service struct {
	database *gorm.DB
}

// New 创建账号服务。
func New(database *gorm.DB) *Service {
	return &Service{database: database}
}

// Register 创建账号（不设置安全问题）；保留给只需要账号的调用方与测试。
func (service *Service) Register(displayName string, password string) (*store.User, error) {
	return service.RegisterWithSecurity(displayName, password, "", "")
}

// RegisterWithSecurity 创建账号，可选地同时设置安全问题与答案。
// 昵称唯一，密码与答案都只存 bcrypt 哈希。
func (service *Service) RegisterWithSecurity(
	displayName string,
	password string,
	securityQuestion string,
	securityAnswer string,
) (*store.User, error) {
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

	question, answerHash, err := prepareSecurity(securityQuestion, securityAnswer)
	if err != nil {
		return nil, err
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

	user := store.User{
		DisplayName:        name,
		PasswordHash:       hashedPassword,
		SecurityQuestion:   question,
		SecurityAnswerHash: answerHash,
	}
	if err := service.database.Create(&user).Error; err != nil {
		// 两个人同时抢同一个昵称时，靠唯一索引兜底。
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrNameTaken
		}
		return nil, err
	}
	return &user, nil
}

// prepareSecurity 校验安全问题与答案，并返回可直接落库的值。
// 问题和答案必须同时为空（不启用自助重置）或同时填写。
func prepareSecurity(question string, answer string) (string, string, error) {
	trimmedQuestion := strings.TrimSpace(question)
	trimmedAnswer := strings.TrimSpace(answer)
	if trimmedQuestion == "" && trimmedAnswer == "" {
		return "", "", nil
	}
	if trimmedQuestion == "" || trimmedAnswer == "" {
		return "", "", ErrSecuritySetupIncomplete
	}

	questionLength := utf8.RuneCountInString(trimmedQuestion)
	if questionLength < MinSecurityQuestionLength || questionLength > MaxSecurityQuestionLength {
		return "", "", &InvalidSecurityQuestionError{
			MinLength: MinSecurityQuestionLength,
			MaxLength: MaxSecurityQuestionLength,
		}
	}
	answerLength := utf8.RuneCountInString(trimmedAnswer)
	if answerLength < MinSecurityAnswerLength || len(trimmedAnswer) > MaxSecurityAnswerBytes {
		return "", "", &InvalidSecurityAnswerError{
			MinLength: MinSecurityAnswerLength,
			MaxBytes:  MaxSecurityAnswerBytes,
		}
	}

	answerHash, err := auth.HashSecurityAnswer(trimmedAnswer)
	if err != nil {
		return "", "", err
	}
	return trimmedQuestion, answerHash, nil
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

// ChangePassword 用当前密码换新密码；密码策略与注册时一致。
func (service *Service) ChangePassword(
	userID uint,
	currentPassword string,
	newPassword string,
) (*store.User, error) {
	user, err := service.ByID(userID)
	if err != nil {
		return nil, err
	}
	if !auth.VerifyPassword(user.PasswordHash, currentPassword) {
		return nil, ErrCurrentPasswordIncorrect
	}
	if unmetRules := auth.CheckPasswordPolicy(newPassword); len(unmetRules) > 0 {
		return nil, &WeakPasswordError{Rules: unmetRules}
	}
	if auth.VerifyPassword(user.PasswordHash, newPassword) {
		return nil, ErrPasswordUnchanged
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	if err := service.database.Model(&store.User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword).Error; err != nil {
		return nil, err
	}
	user.PasswordHash = hashedPassword
	return user, nil
}

// SetSecurityQuestion 设置或更新安全问题；需要当前密码，防止会话被劫持后直接改掉找回方式。
func (service *Service) SetSecurityQuestion(
	userID uint,
	currentPassword string,
	question string,
	answer string,
) (*store.User, error) {
	user, err := service.ByID(userID)
	if err != nil {
		return nil, err
	}
	if !auth.VerifyPassword(user.PasswordHash, currentPassword) {
		return nil, ErrCurrentPasswordIncorrect
	}

	preparedQuestion, answerHash, err := prepareSecurity(question, answer)
	if err != nil {
		return nil, err
	}
	if err := service.database.Model(&store.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"security_question":    preparedQuestion,
			"security_answer_hash": answerHash,
		}).Error; err != nil {
		return nil, err
	}
	user.SecurityQuestion = preparedQuestion
	user.SecurityAnswerHash = answerHash
	return user, nil
}

// QuestionFor 返回账号设置的安全问题；账号不存在或没设置时统一返回不可用。
func (service *Service) QuestionFor(displayName string) (string, error) {
	var user store.User
	err := service.database.
		Select("security_question").
		Where("display_name = ?", strings.TrimSpace(displayName)).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrPasswordResetUnavailable
		}
		return "", err
	}
	if user.SecurityQuestion == "" {
		return "", ErrPasswordResetUnavailable
	}
	return user.SecurityQuestion, nil
}

// ResetPassword 用安全问题答案重置密码。
// 账号不存在、没设置问题、答案不对都返回同一个错误，避免泄露账号是否存在。
func (service *Service) ResetPassword(
	displayName string,
	securityAnswer string,
	newPassword string,
) (*store.User, error) {
	var user store.User
	err := service.database.
		Where("display_name = ?", strings.TrimSpace(displayName)).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 账号不存在也走一次同样代价的校验，避免用响应时间探测账号。
			auth.VerifyAbsentAccount(securityAnswer)
			return nil, ErrPasswordResetFailed
		}
		return nil, err
	}
	if user.SecurityQuestion == "" ||
		!auth.VerifySecurityAnswer(user.SecurityAnswerHash, securityAnswer) {
		return nil, ErrPasswordResetFailed
	}
	if unmetRules := auth.CheckPasswordPolicy(newPassword); len(unmetRules) > 0 {
		return nil, &WeakPasswordError{Rules: unmetRules}
	}
	if auth.VerifyPassword(user.PasswordHash, newPassword) {
		return nil, ErrPasswordUnchanged
	}

	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	if err := service.database.Model(&store.User{}).
		Where("id = ?", user.ID).
		Update("password_hash", hashedPassword).Error; err != nil {
		return nil, err
	}
	user.PasswordHash = hashedPassword
	return &user, nil
}
