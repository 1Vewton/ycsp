package api

import (
	"context"
	"errors"
	"log"
	"net"
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
}

func newAuthHandlers(options AuthOptions) authHandlers {
	return authHandlers{
		users:        options.Users,
		sessions:     options.Sessions,
		guard:        options.LoginGuard,
		lifetime:     options.SessionLifetime,
		secureCookie: options.SecureCookies,
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
}

type credentialsRequest struct {
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// userPayload 是账号的对外形状：绝不包含密码哈希。
type userPayload struct {
	ID          uint      `json:"id"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

func userToPayload(user *store.User) userPayload {
	return userPayload{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt.UTC(),
	}
}

// register 学生自助注册：成功后直接登录（下发会话 Cookie）。
func (handlers authHandlers) register(context *gin.Context) {
	var request credentialsRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
		return
	}

	user, err := handlers.users.Register(request.DisplayName, request.Password)
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
	address := clientAddress(context)
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
			log.Printf("%s reset login counter: %v", ServiceName, err)
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
		log.Printf("%s login guard check: %v", ServiceName, err)
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
		log.Printf("%s login guard record: %v", ServiceName, err)
		return 0, false
	}
	return retryAfter, retryAfter > 0
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

// clientAddress 取直连地址（不带端口）。这里刻意不看 X-Forwarded-For：
// 那个头可以随便伪造，用它统计等于把限流交给攻击者；部署到反向代理后需要另配可信代理。
func clientAddress(context *gin.Context) string {
	host, _, err := net.SplitHostPort(context.Request.RemoteAddr)
	if err != nil {
		return context.Request.RemoteAddr
	}
	return host
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
