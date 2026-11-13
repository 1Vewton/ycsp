package users_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/auth"
	"github.com/1Vewton/ycsp/backend/internal/testutil"
	"github.com/1Vewton/ycsp/backend/internal/users"
)

// 满足密码策略：至少 8 位、含字母、数字、特殊符号。
const validPassword = "ycsp2026!"

func TestRegisterStoresOnlyAHash(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	user, err := service.Register("  小明  ", validPassword)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if user.DisplayName != "小明" {
		t.Errorf("display name = %q, want the trimmed input", user.DisplayName)
	}
	if user.PasswordHash == "" || strings.Contains(user.PasswordHash, validPassword) {
		t.Error("password hash is empty or contains the plaintext password")
	}
	if !auth.VerifyPassword(user.PasswordHash, validPassword) {
		t.Error("stored hash does not verify the password")
	}
}

func TestRegisterRejectsDuplicateNames(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	if _, err := service.Register("小明", validPassword); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	if _, err := service.Register("小明", validPassword); !errors.Is(err, users.ErrNameTaken) {
		t.Errorf("error = %v, want ErrNameTaken", err)
	}
}

func TestRegisterRejectsNamesOutOfRange(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	for _, name := range []string{"", "  ", "小", strings.Repeat("小", 33)} {
		_, err := service.Register(name, validPassword)
		var invalidName *users.InvalidNameError
		if !errors.As(err, &invalidName) {
			t.Errorf("name %q: error = %v, want InvalidNameError", name, err)
		}
	}
}

func TestRegisterRejectsWeakPasswords(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	_, err := service.Register("小明", "abc")
	var weakPassword *users.WeakPasswordError
	if !errors.As(err, &weakPassword) {
		t.Fatalf("error = %v, want WeakPasswordError", err)
	}

	unmetRules := map[auth.PasswordRule]bool{}
	for _, rule := range weakPassword.Rules {
		unmetRules[rule] = true
	}
	for _, wanted := range []auth.PasswordRule{
		auth.PasswordRuleTooShort,
		auth.PasswordRuleMissingDigit,
		auth.PasswordRuleMissingSymbol,
	} {
		if !unmetRules[wanted] {
			t.Errorf("rules = %v, missing %q", weakPassword.Rules, wanted)
		}
	}
}

func TestAuthenticateAcceptsTheRightPassword(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))
	registered, err := service.Register("小明", validPassword)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	user, err := service.Authenticate("  小明 ", validPassword)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.ID != registered.ID {
		t.Errorf("user id = %d, want %d", user.ID, registered.ID)
	}
}

func TestAuthenticateRejectsWrongOrUnknownCredentials(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))
	if _, err := service.Register("小明", validPassword); err != nil {
		t.Fatalf("Register: %v", err)
	}

	attempts := map[string]struct{ name, password string }{
		"wrong password": {"小明", "ycsp2026?"},
		"unknown name":   {"小红", validPassword},
		"empty password": {"小明", ""},
	}
	for description, attempt := range attempts {
		if _, err := service.Authenticate(attempt.name, attempt.password); !errors.Is(
			err, users.ErrInvalidCredentials,
		) {
			t.Errorf("%s: error = %v, want ErrInvalidCredentials", description, err)
		}
	}
}

func TestByIDReportsMissingUsers(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	if _, err := service.ByID(9999); !errors.Is(err, users.ErrUserNotFound) {
		t.Errorf("error = %v, want ErrUserNotFound", err)
	}
}

const (
	securityQuestion = "你最喜欢的元素是什么？"
	securityAnswer   = "钠Na"
	newPassword      = "ycsp2027!"
)

func TestRegisterWithSecurityStoresOnlyAnswerHash(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	user, err := service.RegisterWithSecurity(
		"小明", validPassword, "  你最喜欢的元素是什么？  ", "  钠Na  ",
	)
	if err != nil {
		t.Fatalf("RegisterWithSecurity: %v", err)
	}

	if user.SecurityQuestion != "你最喜欢的元素是什么？" {
		t.Errorf("security question = %q, want the trimmed input", user.SecurityQuestion)
	}
	if user.SecurityAnswerHash == "" || strings.Contains(user.SecurityAnswerHash, "钠Na") {
		t.Error("security answer hash is empty or contains the plaintext answer")
	}
	if !auth.VerifySecurityAnswer(user.SecurityAnswerHash, " 钠NA ") {
		t.Error("stored hash does not verify a normalized answer")
	}
}

func TestRegisterRejectsIncompleteOrInvalidSecuritySetup(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))

	_, err := service.RegisterWithSecurity("小明", validPassword, securityQuestion, "")
	if !errors.Is(err, users.ErrSecuritySetupIncomplete) {
		t.Errorf("missing answer: error = %v, want ErrSecuritySetupIncomplete", err)
	}

	_, err = service.RegisterWithSecurity("小明", validPassword, "", securityAnswer)
	if !errors.Is(err, users.ErrSecuritySetupIncomplete) {
		t.Errorf("missing question: error = %v, want ErrSecuritySetupIncomplete", err)
	}

	_, err = service.RegisterWithSecurity("小明", validPassword, "短", securityAnswer)
	var invalidQuestion *users.InvalidSecurityQuestionError
	if !errors.As(err, &invalidQuestion) {
		t.Errorf("short question: error = %v, want InvalidSecurityQuestionError", err)
	}

	_, err = service.RegisterWithSecurity("小明", validPassword, securityQuestion, "短")
	var invalidAnswer *users.InvalidSecurityAnswerError
	if !errors.As(err, &invalidAnswer) {
		t.Errorf("short answer: error = %v, want InvalidSecurityAnswerError", err)
	}
}

func TestChangePassword(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))
	registered, err := service.Register("小明", validPassword)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	updated, err := service.ChangePassword(registered.ID, validPassword, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if !auth.VerifyPassword(updated.PasswordHash, newPassword) {
		t.Error("new password does not verify against the stored hash")
	}
	if _, err := service.Authenticate("小明", validPassword); !errors.Is(
		err, users.ErrInvalidCredentials,
	) {
		t.Errorf("old password: error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := service.Authenticate("小明", newPassword); err != nil {
		t.Errorf("new password: Authenticate error = %v", err)
	}

	if _, err := service.ChangePassword(registered.ID, validPassword, newPassword); !errors.Is(
		err, users.ErrCurrentPasswordIncorrect,
	) {
		t.Errorf("wrong current password: error = %v, want ErrCurrentPasswordIncorrect", err)
	}
	if _, err := service.ChangePassword(registered.ID, newPassword, newPassword); !errors.Is(
		err, users.ErrPasswordUnchanged,
	) {
		t.Errorf("same password: error = %v, want ErrPasswordUnchanged", err)
	}
	var weakPassword *users.WeakPasswordError
	if _, err := service.ChangePassword(registered.ID, newPassword, "abc"); !errors.As(
		err, &weakPassword,
	) {
		t.Errorf("weak password: error = %v, want WeakPasswordError", err)
	}
}

func TestSetSecurityQuestionAndResetPassword(t *testing.T) {
	service := users.New(testutil.OpenDatabase(t))
	registered, err := service.Register("小明", validPassword)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	updated, err := service.SetSecurityQuestion(
		registered.ID, validPassword, securityQuestion, securityAnswer,
	)
	if err != nil {
		t.Fatalf("SetSecurityQuestion: %v", err)
	}
	if updated.SecurityQuestion != securityQuestion {
		t.Errorf("security question = %q, want %q", updated.SecurityQuestion, securityQuestion)
	}
	if !auth.VerifySecurityAnswer(updated.SecurityAnswerHash, securityAnswer) {
		t.Error("security answer hash does not verify the answer")
	}
	if _, err := service.SetSecurityQuestion(
		registered.ID, "wrong-password", securityQuestion, securityAnswer,
	); !errors.Is(err, users.ErrCurrentPasswordIncorrect) {
		t.Errorf("wrong current password: error = %v, want ErrCurrentPasswordIncorrect", err)
	}

	question, err := service.QuestionFor("  小明 ")
	if err != nil {
		t.Fatalf("QuestionFor: %v", err)
	}
	if question != securityQuestion {
		t.Errorf("question = %q, want %q", question, securityQuestion)
	}
	if _, err := service.QuestionFor("小红"); !errors.Is(
		err, users.ErrPasswordResetUnavailable,
	) {
		t.Errorf("unknown account: error = %v, want ErrPasswordResetUnavailable", err)
	}

	if _, err := service.ResetPassword("小明", "wrong answer", newPassword); !errors.Is(
		err, users.ErrPasswordResetFailed,
	) {
		t.Errorf("wrong answer: error = %v, want ErrPasswordResetFailed", err)
	}
	if _, err := service.ResetPassword("小红", securityAnswer, newPassword); !errors.Is(
		err, users.ErrPasswordResetFailed,
	) {
		t.Errorf("unknown account reset: error = %v, want ErrPasswordResetFailed", err)
	}

	if _, err := service.ResetPassword(" 小明 ", " 钠NA ", newPassword); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if _, err := service.Authenticate("小明", validPassword); !errors.Is(
		err, users.ErrInvalidCredentials,
	) {
		t.Errorf("old password after reset: error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := service.Authenticate("小明", newPassword); err != nil {
		t.Errorf("new password after reset: Authenticate error = %v", err)
	}
	if _, err := service.ResetPassword("小明", securityAnswer, newPassword); !errors.Is(
		err, users.ErrPasswordUnchanged,
	) {
		t.Errorf("same password reset: error = %v, want ErrPasswordUnchanged", err)
	}
}
