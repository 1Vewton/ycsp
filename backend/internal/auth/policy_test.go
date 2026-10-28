package auth_test

import (
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/auth"
)

func containsRule(rules []auth.PasswordRule, wanted auth.PasswordRule) bool {
	for _, rule := range rules {
		if rule == wanted {
			return true
		}
	}
	return false
}

func TestCheckPasswordPolicyAcceptsAStrongPassword(t *testing.T) {
	if unmetRules := auth.CheckPasswordPolicy("ycsp2026!"); len(unmetRules) != 0 {
		t.Fatalf("unmet rules = %v, want none", unmetRules)
	}
}

func TestCheckPasswordPolicyListsEveryUnmetRule(t *testing.T) {
	// 只有小写字母：太短、缺数字、缺特殊符号。
	unmetRules := auth.CheckPasswordPolicy("abc")

	wanted := []auth.PasswordRule{
		auth.PasswordRuleTooShort,
		auth.PasswordRuleMissingDigit,
		auth.PasswordRuleMissingSymbol,
	}
	if len(unmetRules) != len(wanted) {
		t.Fatalf("unmet rules = %v, want %v", unmetRules, wanted)
	}
	for index, rule := range wanted {
		if unmetRules[index] != rule {
			t.Errorf("rule[%d] = %q, want %q", index, unmetRules[index], rule)
		}
	}
}

func TestCheckPasswordPolicyRequiresAnAsciiLetter(t *testing.T) {
	// 中文不算“英文字母”：即使有数字和符号，仍然要提示补字母。
	unmetRules := auth.CheckPasswordPolicy("密码密码密码1!")

	if !containsRule(unmetRules, auth.PasswordRuleMissingLetter) {
		t.Errorf("unmet rules = %v, want %q", unmetRules, auth.PasswordRuleMissingLetter)
	}
}

func TestCheckPasswordPolicyDoesNotCountSpacesAsSymbols(t *testing.T) {
	unmetRules := auth.CheckPasswordPolicy("abc 1234")

	if !containsRule(unmetRules, auth.PasswordRuleMissingSymbol) {
		t.Errorf("unmet rules = %v, want %q", unmetRules, auth.PasswordRuleMissingSymbol)
	}
}

func TestCheckPasswordPolicyRejectsPasswordsLongerThanBcrypt(t *testing.T) {
	// 恰好 72 字节可以；73 字节必须报错（bcrypt 会静默截断）。
	exactlyAtLimit := strings.Repeat("a", 70) + "1!" // 72 字节
	if unmetRules := auth.CheckPasswordPolicy(exactlyAtLimit); len(unmetRules) != 0 {
		t.Errorf("72 bytes should pass, got %v", unmetRules)
	}

	tooLong := strings.Repeat("a", 71) + "1!" // 73 字节
	if unmetRules := auth.CheckPasswordPolicy(tooLong); !containsRule(unmetRules, auth.PasswordRuleTooLong) {
		t.Errorf("unmet rules = %v, want %q", unmetRules, auth.PasswordRuleTooLong)
	}
}
