package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/ratelimit"
	"github.com/1Vewton/ycsp/backend/internal/session"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
	"github.com/1Vewton/ycsp/backend/internal/testutil"
	"github.com/1Vewton/ycsp/backend/internal/users"
)

// 满足密码策略：至少 8 位、含字母、数字、特殊符号。
const validPassword = "ycsp2026!"

// newAuthRouter 建一个带账号接口的路由（会话与限流都放内存；Redis 的行为在各自包里测）。
// 默认阈值放宽到“账号 5 次、地址 100 次”，避免和限流测试互相干扰。
func newAuthRouter(t *testing.T) (http.Handler, *gorm.DB) {
	t.Helper()

	return newAuthRouterWithPolicy(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 5,
		MaxAttemptsPerAddress: 100,
		Window:                time.Minute,
	})
}

func newAuthRouterWithPolicy(t *testing.T, policy ratelimit.Policy) (http.Handler, *gorm.DB) {
	t.Helper()

	database := testutil.OpenDatabase(t)
	guard, err := ratelimit.NewGuard(ratelimit.NewMemoryCounter(), policy)
	if err != nil {
		t.Fatalf("NewGuard: %v", err)
	}

	router := api.NewRouter(api.Options{
		ComputeBaseURL: "http://127.0.0.1:1",
		Auth: api.AuthOptions{
			Users:           users.New(database),
			Sessions:        session.NewMemoryStore(),
			SessionLifetime: time.Hour,
			LoginGuard:      guard,
		},
	})
	return router, database
}

func performAuthRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
	cookie *http.Cookie,
) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not JSON: %v (body=%q)", err, recorder.Body.String())
	}
	return recorder, payload
}

func sessionCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "ycsp_session" {
			return cookie
		}
	}
	t.Fatal("response set no session cookie")
	return nil
}

// registerUser 注册一个账号并返回登录后的 Cookie。
func registerUser(t *testing.T, router http.Handler, displayName string) *http.Cookie {
	t.Helper()

	recorder, _ := performAuthRequest(
		t, router, http.MethodPost, "/api/users",
		`{"display_name":"`+displayName+`","password":"`+validPassword+`"}`, nil,
	)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	return sessionCookie(t, recorder)
}

func TestRegisterCreatesAccountAndSignsIn(t *testing.T) {
	router, _ := newAuthRouter(t)

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/users",
		`{"display_name":" 小明 ","password":"`+validPassword+`"}`, nil,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if payload["code"] != codes.UserRegistered {
		t.Errorf("code = %v, want %q", payload["code"], codes.UserRegistered)
	}

	user, ok := payload["user"].(map[string]any)
	if !ok {
		t.Fatalf("payload has no user: %v", payload)
	}
	if user["display_name"] != "小明" {
		t.Errorf("display_name = %v, want the trimmed input", user["display_name"])
	}
	if _, leaked := user["password_hash"]; leaked {
		t.Error("user payload must not contain the password hash")
	}

	cookie := sessionCookie(t, recorder)
	if !cookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}

	meRecorder, mePayload := performAuthRequest(t, router, http.MethodGet, "/api/users/me", "", cookie)
	if meRecorder.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d", meRecorder.Code, http.StatusOK)
	}
	if mePayload["code"] != codes.UserOK {
		t.Errorf("me code = %v, want %q", mePayload["code"], codes.UserOK)
	}
	if mePayload["user"].(map[string]any)["display_name"] != "小明" {
		t.Errorf("me user = %v", mePayload["user"])
	}
}

func TestRegisterRejectsWeakPasswordListingEveryRule(t *testing.T) {
	router, _ := newAuthRouter(t)

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/users",
		`{"display_name":"小明","password":"abc"}`, nil,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if payload["code"] != codes.PasswordTooWeak {
		t.Errorf("code = %v, want %q", payload["code"], codes.PasswordTooWeak)
	}

	params, ok := payload["params"].(map[string]any)
	if !ok {
		t.Fatalf("params = %v", payload["params"])
	}
	if params["min_length"] != float64(8) {
		t.Errorf("min_length = %v, want 8", params["min_length"])
	}
	rules, ok := params["rules"].([]any)
	if !ok {
		t.Fatalf("rules = %v", params["rules"])
	}
	ruleNames := map[string]bool{}
	for _, rule := range rules {
		ruleNames[rule.(string)] = true
	}
	for _, wanted := range []string{
		"password_rule_too_short",
		"password_rule_missing_digit",
		"password_rule_missing_symbol",
	} {
		if !ruleNames[wanted] {
			t.Errorf("rules = %v, missing %q", rules, wanted)
		}
	}
}

func TestRegisterRejectsTakenName(t *testing.T) {
	router, _ := newAuthRouter(t)
	registerUser(t, router, "小明")

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/users",
		`{"display_name":"小明","password":"`+validPassword+`"}`, nil,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	if payload["code"] != codes.UserNameTaken {
		t.Errorf("code = %v, want %q", payload["code"], codes.UserNameTaken)
	}
}

func TestRegisterRejectsInvalidName(t *testing.T) {
	router, _ := newAuthRouter(t)

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/users",
		`{"display_name":" ","password":"`+validPassword+`"}`, nil,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if payload["code"] != codes.UserNameInvalid {
		t.Errorf("code = %v, want %q", payload["code"], codes.UserNameInvalid)
	}
	params, ok := payload["params"].(map[string]any)
	if !ok || params["min_length"] != float64(2) || params["max_length"] != float64(32) {
		t.Errorf("params = %v, want min_length 2 / max_length 32", payload["params"])
	}
}

func TestRegisterRejectsInvalidJSON(t *testing.T) {
	router, _ := newAuthRouter(t)

	recorder, payload := performAuthRequest(t, router, http.MethodPost, "/api/users", "not json", nil)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if payload["code"] != codes.RequestInvalidBody {
		t.Errorf("code = %v, want %q", payload["code"], codes.RequestInvalidBody)
	}
}

func TestSignInSetsASessionCookie(t *testing.T) {
	router, _ := newAuthRouter(t)
	registerUser(t, router, "小明")

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions",
		`{"display_name":"小明","password":"`+validPassword+`"}`, nil,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if payload["code"] != codes.SignedIn {
		t.Errorf("code = %v, want %q", payload["code"], codes.SignedIn)
	}
	cookie := sessionCookie(t, recorder)

	meRecorder, _ := performAuthRequest(t, router, http.MethodGet, "/api/users/me", "", cookie)
	if meRecorder.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d", meRecorder.Code, http.StatusOK)
	}
}

func TestSignInRejectsWrongOrUnknownCredentials(t *testing.T) {
	router, _ := newAuthRouter(t)
	registerUser(t, router, "小明")

	attempts := map[string]string{
		"wrong password": `{"display_name":"小明","password":"ycsp2026?"}`,
		"unknown name":   `{"display_name":"小红","password":"` + validPassword + `"}`,
	}
	for description, body := range attempts {
		recorder, payload := performAuthRequest(t, router, http.MethodPost, "/api/sessions", body, nil)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want %d", description, recorder.Code, http.StatusUnauthorized)
		}
		if payload["code"] != codes.InvalidCredentials {
			t.Errorf("%s: code = %v, want %q", description, payload["code"], codes.InvalidCredentials)
		}
	}
}

func TestMeRequiresASession(t *testing.T) {
	router, _ := newAuthRouter(t)

	anonymousRecorder, anonymousPayload := performAuthRequest(
		t, router, http.MethodGet, "/api/users/me", "", nil,
	)
	if anonymousRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want %d", anonymousRecorder.Code, http.StatusUnauthorized)
	}
	if anonymousPayload["code"] != codes.NotSignedIn {
		t.Errorf("anonymous code = %v, want %q", anonymousPayload["code"], codes.NotSignedIn)
	}

	staleCookie := &http.Cookie{Name: "ycsp_session", Value: "stale"}
	staleRecorder, stalePayload := performAuthRequest(
		t, router, http.MethodGet, "/api/users/me", "", staleCookie,
	)
	if staleRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("stale status = %d, want %d", staleRecorder.Code, http.StatusUnauthorized)
	}
	if stalePayload["code"] != codes.NotSignedIn {
		t.Errorf("stale code = %v, want %q", stalePayload["code"], codes.NotSignedIn)
	}
}

func TestSignOutClearsTheSession(t *testing.T) {
	router, _ := newAuthRouter(t)
	cookie := registerUser(t, router, "小明")

	recorder, payload := performAuthRequest(t, router, http.MethodDelete, "/api/sessions", "", cookie)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if payload["code"] != codes.SignedOut {
		t.Errorf("code = %v, want %q", payload["code"], codes.SignedOut)
	}

	clearedCookie := sessionCookie(t, recorder)
	if clearedCookie.MaxAge >= 0 {
		t.Error("sign out should expire the session cookie")
	}

	meRecorder, _ := performAuthRequest(t, router, http.MethodGet, "/api/users/me", "", cookie)
	if meRecorder.Code != http.StatusUnauthorized {
		t.Errorf("me after sign out = %d, want %d", meRecorder.Code, http.StatusUnauthorized)
	}
}

func TestSignInLocksAfterRepeatedFailures(t *testing.T) {
	router, _ := newAuthRouterWithPolicy(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 3,
		MaxAttemptsPerAddress: 100,
		Window:                time.Minute,
	})
	registerUser(t, router, "小明")

	wrongCredentials := `{"display_name":"小明","password":"ycsp2026?"}`
	for attempt := 1; attempt <= 2; attempt++ {
		recorder, payload := performAuthRequest(
			t, router, http.MethodPost, "/api/sessions", wrongCredentials, nil,
		)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, http.StatusUnauthorized)
		}
		if payload["code"] != codes.InvalidCredentials {
			t.Errorf("attempt %d code = %v, want %q", attempt, payload["code"], codes.InvalidCredentials)
		}
	}

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions", wrongCredentials, nil,
	)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if payload["code"] != codes.TooManyAttempts {
		t.Errorf("code = %v, want %q", payload["code"], codes.TooManyAttempts)
	}

	params, ok := payload["params"].(map[string]any)
	if !ok {
		t.Fatalf("params = %v", payload["params"])
	}
	if params["retry_after_seconds"].(float64) <= 0 {
		t.Errorf("retry_after_seconds = %v, want > 0", params["retry_after_seconds"])
	}
	if params["retry_after_minutes"].(float64) < 1 {
		t.Errorf("retry_after_minutes = %v, want >= 1", params["retry_after_minutes"])
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Error("response should carry a Retry-After header")
	}
}

func TestBlockedSignInRejectsEvenTheCorrectPassword(t *testing.T) {
	router, _ := newAuthRouterWithPolicy(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 1,
		MaxAttemptsPerAddress: 100,
		Window:                time.Minute,
	})
	registerUser(t, router, "小明")

	// 第一次失败就达到上限，响应直接变成 429。
	lockedRecorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions",
		`{"display_name":"小明","password":"ycsp2026?"}`, nil,
	)
	if lockedRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("first failure status = %d, want %d", lockedRecorder.Code, http.StatusTooManyRequests)
	}
	if payload["code"] != codes.TooManyAttempts {
		t.Errorf("code = %v, want %q", payload["code"], codes.TooManyAttempts)
	}

	// 锁定期间即使密码正确也不能登录（否则限流等于没做）。
	correctRecorder, correctPayload := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions",
		`{"display_name":"小明","password":"`+validPassword+`"}`, nil,
	)
	if correctRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("correct password status = %d, want %d", correctRecorder.Code, http.StatusTooManyRequests)
	}
	if correctPayload["code"] != codes.TooManyAttempts {
		t.Errorf("code = %v, want %q", correctPayload["code"], codes.TooManyAttempts)
	}
}

func TestSuccessfulSignInResetsTheAccountCounter(t *testing.T) {
	router, _ := newAuthRouterWithPolicy(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 3,
		MaxAttemptsPerAddress: 100,
		Window:                time.Minute,
	})
	registerUser(t, router, "小明")

	wrongCredentials := `{"display_name":"小明","password":"ycsp2026?"}`
	for attempt := 1; attempt <= 2; attempt++ {
		recorder, _ := performAuthRequest(
			t, router, http.MethodPost, "/api/sessions", wrongCredentials, nil,
		)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, http.StatusUnauthorized)
		}
	}

	successRecorder, _ := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions",
		`{"display_name":"小明","password":"`+validPassword+`"}`, nil,
	)
	if successRecorder.Code != http.StatusOK {
		t.Fatalf("success status = %d, want %d", successRecorder.Code, http.StatusOK)
	}

	// 计数已清零：还能再错 2 次才到上限，说明成功登录把失败记录抹掉了。
	for attempt := 1; attempt <= 2; attempt++ {
		recorder, _ := performAuthRequest(
			t, router, http.MethodPost, "/api/sessions", wrongCredentials, nil,
		)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d after reset status = %d, want %d", attempt, recorder.Code, http.StatusUnauthorized)
		}
	}
}

func TestSignInLocksByAddressAcrossAccounts(t *testing.T) {
	router, _ := newAuthRouterWithPolicy(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 5,
		MaxAttemptsPerAddress: 2,
		Window:                time.Minute,
	})
	registerUser(t, router, "小明")
	registerUser(t, router, "小红")

	firstRecorder, _ := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions",
		`{"display_name":"小明","password":"ycsp2026?"}`, nil,
	)
	if firstRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("first attempt status = %d, want %d", firstRecorder.Code, http.StatusUnauthorized)
	}

	// 换一个昵称也没用：同一地址的第 2 次失败就触发限制。
	secondRecorder, secondPayload := performAuthRequest(
		t, router, http.MethodPost, "/api/sessions",
		`{"display_name":"小红","password":"ycsp2026?"}`, nil,
	)
	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second attempt status = %d, want %d", secondRecorder.Code, http.StatusTooManyRequests)
	}
	if secondPayload["code"] != codes.TooManyAttempts {
		t.Errorf("code = %v, want %q", secondPayload["code"], codes.TooManyAttempts)
	}
}

func TestTaskCreatedWhileSignedInRecordsTheOwner(t *testing.T) {
	database := testutil.OpenDatabase(t)
	router := api.NewRouter(api.Options{
		ComputeBaseURL: "http://127.0.0.1:1",
		Tasks:          tasks.New(database, successPredictor()),
		Auth: api.AuthOptions{
			Users:           users.New(database),
			Sessions:        session.NewMemoryStore(),
			SessionLifetime: time.Hour,
		},
	})
	cookie := registerUser(t, router, "小明")

	recorder, payload := performAuthRequest(
		t, router, http.MethodPost, "/api/tasks", `{"formula":"NaCl"}`, cookie,
	)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	taskID := int(payload["task"].(map[string]any)["id"].(float64))
	var stored store.PredictionTask
	if err := database.First(&stored, taskID).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if stored.UserID == nil {
		t.Fatal("task created by a signed-in user has no user_id")
	}

	var owner store.User
	if err := database.First(&owner, *stored.UserID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if owner.DisplayName != "小明" {
		t.Errorf("task owner = %q, want 小明", owner.DisplayName)
	}
}
