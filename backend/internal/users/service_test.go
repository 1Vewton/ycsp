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
