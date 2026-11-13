package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/auth"
	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/ratelimit"
	"github.com/1Vewton/ycsp/backend/internal/session"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/traceid"
	"github.com/1Vewton/ycsp/backend/internal/users"
)

// sessionCookieName 是保存会话编号的 Cookie 名。
const sessionCookieName = "ycsp_session"

// secondsPerMinute 用于把等待时间换算成“约几分钟”给学生看。
const secondsPerMinute = 60

// authHandlers 承载账号接口的依赖；Users 或 Sessions 为空表示没启用账号体系。
type authHandlers struct {
	users        *users.Service
	sessions     session.Store
	guard        *ratelimit.Guard
	lifetime     time.Duration
	secureCookie bool
	addresses    clientAddressResolver
}

func newAuthHandlers(options AuthOptions) authHandlers {
	return authHandlers{
		users:        options.Users,
		sessions:     options.Sessions,
		guard:        options.LoginGuard,
		lifetime:     options.SessionLifetime,
		secureCookie: options.SecureCookies,
		addresses:    newClientAddressResolver(options.TrustedProxies),
	}
}

func (handlers authHandlers) enabled() bool {
	return handlers.users != nil && handlers.sessions != nil
}

func registerAuthRoutes(router *gin.Engine, handlers authHandlers) {
	router.POST("/api/users", handlers.register)
	router.POST("/api/sessions", handlers.signIn)
	router.DELETE("/api/sessions", handlers.signOut)
	router.GET("/api/users/me", handlers.me)
	router.PUT("/api/users/me/password", handlers.changePassword)
	router.PUT("/api/users/me/security-question", handlers.setSecurityQuestion)
	router.POST("/api/password-resets/question", handlers.passwordResetQuestion)
	router.POST("/api/password-resets", handlers.resetPassword)
}

type credentialsRequest struct {
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// registrationRequest 比登录多两个可选字段：安全问题与答案。
// 两者要么都填（启用自助重置），要么都留空。
type registrationRequest struct {
	DisplayName      string `json:"display_name"`
	Password         string `json:"password"`
	SecurityQuestion string `json:"security_question"`
	SecurityAnswer   string `json:"security_answer"`
}

// passwordChangeRequest 是登录后修改密码的请求体。
type passwordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// securityQuestionRequest 是登录后设置或更新安全问题的请求体。
type securityQuestionRequest struct {
	CurrentPassword  string `json:"current_password"`
	SecurityQuestion string `json:"security_question"`
	SecurityAnswer   string `json:"security_answer"`
}

// passwordResetQuestionRequest 是忘记密码时按昵称取安全问题的请求体。
type passwordResetQuestionRequest struct {
	DisplayName string `json:"display_name"`
}

// passwordResetRequest 是用安全问题答案重置密码的请求体。
type passwordResetRequest struct {
	DisplayName    string `json:"display_name"`
	SecurityAnswer string `json:"security_answer"`
	NewPassword    string `json:"new_password"`
}

// userPayload 是账号的对外形状：绝不包含密码哈希或安全答案哈希。
type userPayload struct {
	ID               uint      `json:"id"`
	DisplayName      string    `json:"display_name"`
	SecurityQuestion string    `json:"security_question"`
	CreatedAt        time.Time `json:"created_at"`
}

func userToPayload(user *store.User) userPayload {
	return userPayload{
		ID:               user.ID,
		DisplayName:      user.DisplayName,
		SecurityQuestion: user.SecurityQuestion,
		CreatedAt:        user.CreatedAt.UTC(),
	}
}

// register 学生自助注册：成功后直接登录（下发会话 Cookie）。
func (handlers authHandlers) register(context *gin.Context) {
	var request registrationRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	user, err := handlers.users.RegisterWithSecurity(
		request.DisplayName,
		request.Password,
		request.SecurityQuestion,
		request.SecurityAnswer,
	)
	if err != nil {
		handlers.writeAccountError(context, err)
		return
	}
	if !handlers.startSession(context, user.ID) {
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.UserRegistered,
		"user":    userToPayload(user),
	})
}

// signIn 校验昵称与密码，成功后下发会话 Cookie。
// 连续失败达到阈值时先返回 429，避免暴力猜密码（见 internal/ratelimit）。
func (handlers authHandlers) signIn(context *gin.Context) {
	var request credentialsRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	account := strings.TrimSpace(request.DisplayName)
	address := handlers.addresses.Address(context.Request)
	requestContext := context.Request.Context()

	if retryAfter, blocked := handlers.blockedFor(requestContext, account, address); blocked {
		writeTooManyAttempts(context, retryAfter)
		return
	}

	user, err := handlers.users.Authenticate(request.DisplayName, request.Password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) {
			if retryAfter, blocked := handlers.recordFailure(
				requestContext, account, address,
			); blocked {
				writeTooManyAttempts(context, retryAfter)
				return
			}
			// 不区分“昵称不存在”和“密码不对”，避免泄露账号是否存在。
			writeError(context, http.StatusUnauthorized, codes.InvalidCredentials, nil)
			return
		}
		writeServerError(context, err)
		return
	}

	if handlers.guard != nil {
		// 登录成功只清这个昵称的计数；清零失败不影响本次登录。
		if err := handlers.guard.ResetAccount(requestContext, account); err != nil {
			log.Printf(
				"%s reset login counter: request_id=%s error=%v",
				ServiceName, traceid.From(requestContext), err,
			)
		}
	}

	if !handlers.startSession(context, user.ID) {
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.SignedIn,
		"user":    userToPayload(user),
	})
}

// blockedFor 判断当前是否已被限流；计数服务出错时放行并记日志——
// 限流坏了也不该把学生挡在门外，代价是这一次少一层保护。
func (handlers authHandlers) blockedFor(
	ctx context.Context, account string, address string,
) (time.Duration, bool) {
	if handlers.guard == nil {
		return 0, false
	}

	retryAfter, err := handlers.guard.Check(ctx, account, address)
	if err != nil {
		log.Printf(
			"%s login guard check: request_id=%s error=%v",
			ServiceName, traceid.From(ctx), err,
		)
		return 0, false
	}
	return retryAfter, retryAfter > 0
}

// recordFailure 记一次失败；返回 true 表示这次失败已经触发限制。
func (handlers authHandlers) recordFailure(
	ctx context.Context, account string, address string,
) (time.Duration, bool) {
	if handlers.guard == nil {
		return 0, false
	}

	retryAfter, err := handlers.guard.RecordFailure(ctx, account, address)
	if err != nil {
		log.Printf(
			"%s login guard record: request_id=%s error=%v",
			ServiceName, traceid.From(ctx), err,
		)
		return 0, false
	}
	return retryAfter, retryAfter > 0
}

// resetFailures 在密码/安全问题校验成功后清掉这个昵称的失败计数。
func (handlers authHandlers) resetFailures(ctx context.Context, account string) {
	if handlers.guard == nil {
		return
	}
	if err := handlers.guard.ResetAccount(ctx, account); err != nil {
		log.Printf(
			"%s reset login counter: request_id=%s error=%v",
			ServiceName, traceid.From(ctx), err,
		)
	}
}

// signOut 结束当前会话；没有会话时也返回成功（退出登录是幂等的）。
func (handlers authHandlers) signOut(context *gin.Context) {
	sessionID, err := context.Cookie(sessionCookieName)
	if err == nil && sessionID != "" {
		if deleteErr := handlers.sessions.Delete(context.Request.Context(), sessionID); deleteErr != nil {
			writeError(context, http.StatusInternalServerError, codes.SessionUnavailable, nil)
			return
		}
	}
	handlers.clearSessionCookie(context)

	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.SignedOut,
	})
}

// me 返回当前登录用户，供前端判断“是否已登录”。
func (handlers authHandlers) me(context *gin.Context) {
	user, ok := handlers.currentUser(context)
	if !ok {
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.UserOK,
		"user":    userToPayload(user),
	})
}

// changePassword 登录后修改密码：需要当前密码，成功后当前会话继续有效。
func (handlers authHandlers) changePassword(context *gin.Context) {
	user, ok := handlers.currentUser(context)
	if !ok {
		return
	}

	var request passwordChangeRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	requestContext := context.Request.Context()
	account := user.DisplayName
	address := handlers.addresses.Address(context.Request)
	if retryAfter, blocked := handlers.blockedFor(requestContext, account, address); blocked {
		writeTooManyAttempts(context, retryAfter)
		return
	}

	updatedUser, err := handlers.users.ChangePassword(
		user.ID, request.CurrentPassword, request.NewPassword,
	)
	if err != nil {
		if errors.Is(err, users.ErrCurrentPasswordIncorrect) {
			retryAfter, blocked := handlers.recordFailure(requestContext, account, address)
			if blocked {
				writeTooManyAttempts(context, retryAfter)
				return
			}
		}
		handlers.writeAccountError(context, err)
		return
	}

	handlers.resetFailures(requestContext, account)
	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.PasswordChanged,
		"user":    userToPayload(updatedUser),
	})
}

// setSecurityQuestion 登录后设置或更新安全问题：需要当前密码。
func (handlers authHandlers) setSecurityQuestion(context *gin.Context) {
	user, ok := handlers.currentUser(context)
	if !ok {
		return
	}

	var request securityQuestionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	requestContext := context.Request.Context()
	account := user.DisplayName
	address := handlers.addresses.Address(context.Request)
	if retryAfter, blocked := handlers.blockedFor(requestContext, account, address); blocked {
		writeTooManyAttempts(context, retryAfter)
		return
	}

	updatedUser, err := handlers.users.SetSecurityQuestion(
		user.ID, request.CurrentPassword, request.SecurityQuestion, request.SecurityAnswer,
	)
	if err != nil {
		if errors.Is(err, users.ErrCurrentPasswordIncorrect) {
			retryAfter, blocked := handlers.recordFailure(requestContext, account, address)
			if blocked {
				writeTooManyAttempts(context, retryAfter)
				return
			}
		}
		handlers.writeAccountError(context, err)
		return
	}

	handlers.resetFailures(requestContext, account)
	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.SecurityQuestionUpdated,
		"user":    userToPayload(updatedUser),
	})
}

// passwordResetQuestion 按昵称返回安全问题；账号不存在或没设置问题统一返回不可用。
func (handlers authHandlers) passwordResetQuestion(context *gin.Context) {
	var request passwordResetQuestionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	question, err := handlers.users.QuestionFor(request.DisplayName)
	if err != nil {
		if errors.Is(err, users.ErrPasswordResetUnavailable) {
			writeError(context, http.StatusBadRequest, codes.PasswordResetUnavailable, nil)
			return
		}
		writeServerError(context, err)
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"service":           ServiceName,
		"status":            "ok",
		"code":              codes.PasswordResetQuestion,
		"security_question": question,
	})
}

// resetPassword 用安全问题答案重置密码；成功后直接登录，避免学生再输一次新密码。
func (handlers authHandlers) resetPassword(context *gin.Context) {
	var request passwordResetRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	requestContext := context.Request.Context()
	account := strings.TrimSpace(request.DisplayName)
	address := handlers.addresses.Address(context.Request)
	if retryAfter, blocked := handlers.blockedFor(requestContext, account, address); blocked {
		writeTooManyAttempts(context, retryAfter)
		return
	}

	user, err := handlers.users.ResetPassword(
		request.DisplayName, request.SecurityAnswer, request.NewPassword,
	)
	if err != nil {
		if errors.Is(err, users.ErrPasswordResetFailed) {
			retryAfter, blocked := handlers.recordFailure(requestContext, account, address)
			if blocked {
				writeTooManyAttempts(context, retryAfter)
				return
			}
		}
		handlers.writeAccountError(context, err)
		return
	}

	handlers.resetFailures(requestContext, account)
	if !handlers.startSession(context, user.ID) {
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.PasswordReset,
		"user":    userToPayload(user),
	})
}

// startSession 创建会话并下发 Cookie；失败时已写好应答。
func (handlers authHandlers) startSession(context *gin.Context, userID uint) bool {
	sessionID, err := handlers.sessions.Create(
		context.Request.Context(), userID, handlers.lifetime,
	)
	if err != nil {
		writeError(context, http.StatusInternalServerError, codes.SessionUnavailable, nil)
		return false
	}

	// HttpOnly 让脚本读不到 Cookie；SameSite=Lax 挡住跨站表单提交；
	// Secure 只在 HTTPS 下开启（本地 http 开发时为 false）。
	context.SetSameSite(http.SameSiteLaxMode)
	context.SetCookie(
		sessionCookieName, sessionID,
		int(handlers.lifetime.Seconds()), "/", "", handlers.secureCookie, true,
	)
	return true
}

func (handlers authHandlers) clearSessionCookie(context *gin.Context) {
	context.SetSameSite(http.SameSiteLaxMode)
	context.SetCookie(sessionCookieName, "", -1, "/", "", handlers.secureCookie, true)
}

// currentUser 取当前登录用户；没有会话或会话失效时写 401 并返回 false。
func (handlers authHandlers) currentUser(context *gin.Context) (*store.User, bool) {
	userID, ok := handlers.sessionUserID(context)
	if !ok {
		return nil, false
	}

	user, err := handlers.users.ByID(userID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			handlers.clearSessionCookie(context)
			writeError(context, http.StatusUnauthorized, codes.NotSignedIn, nil)
			return nil, false
		}
		writeServerError(context, err)
		return nil, false
	}
	return user, true
}

func (handlers authHandlers) sessionUserID(context *gin.Context) (uint, bool) {
	sessionID, err := context.Cookie(sessionCookieName)
	if err != nil || sessionID == "" {
		writeError(context, http.StatusUnauthorized, codes.NotSignedIn, nil)
		return 0, false
	}

	userID, err := handlers.sessions.Lookup(context.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			handlers.clearSessionCookie(context)
			writeError(context, http.StatusUnauthorized, codes.NotSignedIn, nil)
			return 0, false
		}
		writeError(context, http.StatusInternalServerError, codes.SessionUnavailable, nil)
		return 0, false
	}
	return userID, true
}

// optionalUserID 给任务接口用：登录了就记下用户，没登录也照常预测（返回 nil）。
// 这里刻意不写任何错误：账号系统出问题不应该挡住“只想看看结构”的学生。
func (handlers authHandlers) optionalUserID(context *gin.Context) *uint {
	if !handlers.enabled() {
		return nil
	}

	sessionID, err := context.Cookie(sessionCookieName)
	if err != nil || sessionID == "" {
		return nil
	}
	userID, err := handlers.sessions.Lookup(context.Request.Context(), sessionID)
	if err != nil {
		return nil
	}
	return &userID
}

func (handlers authHandlers) writeAccountError(context *gin.Context, err error) {
	var invalidName *users.InvalidNameError
	var weakPassword *users.WeakPasswordError
	var invalidQuestion *users.InvalidSecurityQuestionError
	var invalidAnswer *users.InvalidSecurityAnswerError

	switch {
	case errors.As(err, &invalidName):
		writeError(context, http.StatusBadRequest, codes.UserNameInvalid, gin.H{
			"min_length": invalidName.MinLength,
			"max_length": invalidName.MaxLength,
		})
	case errors.As(err, &weakPassword):
		writeError(context, http.StatusBadRequest, codes.PasswordTooWeak, gin.H{
			"rules":      passwordRulesToStrings(weakPassword.Rules),
			"min_length": auth.MinPasswordLength,
			"max_bytes":  auth.MaxPasswordBytes,
		})
	case errors.As(err, &invalidQuestion):
		writeError(context, http.StatusBadRequest, codes.SecurityQuestionInvalid, gin.H{
			"min_length": invalidQuestion.MinLength,
			"max_length": invalidQuestion.MaxLength,
		})
	case errors.As(err, &invalidAnswer):
		writeError(context, http.StatusBadRequest, codes.SecurityAnswerInvalid, gin.H{
			"min_length": invalidAnswer.MinLength,
			"max_bytes":  invalidAnswer.MaxBytes,
		})
	case errors.Is(err, users.ErrSecuritySetupIncomplete):
		writeError(context, http.StatusBadRequest, codes.SecuritySetupIncomplete, nil)
	case errors.Is(err, users.ErrCurrentPasswordIncorrect):
		writeError(context, http.StatusBadRequest, codes.CurrentPasswordIncorrect, nil)
	case errors.Is(err, users.ErrPasswordUnchanged):
		writeError(context, http.StatusBadRequest, codes.PasswordUnchanged, nil)
	case errors.Is(err, users.ErrPasswordResetUnavailable):
		writeError(context, http.StatusBadRequest, codes.PasswordResetUnavailable, nil)
	case errors.Is(err, users.ErrPasswordResetFailed):
		writeError(context, http.StatusBadRequest, codes.PasswordResetFailed, nil)
	case errors.Is(err, users.ErrNameTaken):
		writeError(context, http.StatusConflict, codes.UserNameTaken, nil)
	default:
		writeServerError(context, err)
	}
}

func passwordRulesToStrings(rules []auth.PasswordRule) []string {
	names := make([]string, 0, len(rules))
	for _, rule := range rules {
		names = append(names, string(rule))
	}
	return names
}

// writeTooManyAttempts 返回 429，并给出还要等多久（同时带标准的 Retry-After 头）。
func writeTooManyAttempts(context *gin.Context, retryAfter time.Duration) {
	seconds := int((retryAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	minutes := (seconds + secondsPerMinute - 1) / secondsPerMinute

	context.Header("Retry-After", strconv.Itoa(seconds))
	writeError(context, http.StatusTooManyRequests, codes.TooManyAttempts, gin.H{
		"retry_after_seconds": seconds,
		"retry_after_minutes": minutes,
	})
}
