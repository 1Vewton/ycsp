package auth

import (
	"unicode"
	"unicode/utf8"
)

// MaxPasswordBytes 是 bcrypt 能处理的密码上限（字节）。
const MaxPasswordBytes = 72

// MinPasswordLength 是密码的最短长度（字符数）。
const MinPasswordLength = 8

// PasswordRule 是一条没被满足的密码规则，直接作为前端语言包的 key 使用。
type PasswordRule string

const (
	PasswordRuleTooShort      PasswordRule = "password_rule_too_short"
	PasswordRuleTooLong       PasswordRule = "password_rule_too_long"
	PasswordRuleMissingLetter PasswordRule = "password_rule_missing_letter"
	PasswordRuleMissingDigit  PasswordRule = "password_rule_missing_digit"
	PasswordRuleMissingSymbol PasswordRule = "password_rule_missing_symbol"
)

// CheckPasswordPolicy 返回所有未满足的规则（一次列全，学生改一次就能通过）。
//
// 策略：至少 8 个字符，且包含英文字母、数字、特殊符号各至少一个。
// “特殊符号”指标点或符号（空格不算）；“英文字母”只认 ASCII，避免中文也算作字母。
func CheckPasswordPolicy(password string) []PasswordRule {
	unmetRules := make([]PasswordRule, 0, 5)

	if utf8.RuneCountInString(password) < MinPasswordLength {
		unmetRules = append(unmetRules, PasswordRuleTooShort)
	}
	if len(password) > MaxPasswordBytes {
		unmetRules = append(unmetRules, PasswordRuleTooLong)
	}

	hasLetter, hasDigit, hasSymbol := false, false, false
	for _, character := range password {
		switch {
		case character < unicode.MaxASCII && unicode.IsLetter(character):
			hasLetter = true
		case character >= '0' && character <= '9':
			hasDigit = true
		case unicode.IsPunct(character) || unicode.IsSymbol(character):
			hasSymbol = true
		}
	}

	if !hasLetter {
		unmetRules = append(unmetRules, PasswordRuleMissingLetter)
	}
	if !hasDigit {
		unmetRules = append(unmetRules, PasswordRuleMissingDigit)
	}
	if !hasSymbol {
		unmetRules = append(unmetRules, PasswordRuleMissingSymbol)
	}
	return unmetRules
}
