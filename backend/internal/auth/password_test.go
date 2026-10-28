package auth_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/auth"
)

// 用一段够长的口令：如果实现真的把明文存了进去，断言会立刻发现。
const plainPassword = "ycsp-教学-2026-strong-passphrase"

func TestHashPasswordStoresNoPlaintext(t *testing.T) {
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if hashedPassword == plainPassword {
		t.Fatal("hash equals the plaintext password")
	}
	if strings.Contains(hashedPassword, plainPassword) {
		t.Fatal("hash contains the plaintext password")
	}
	if !strings.HasPrefix(hashedPassword, "$2") {
		t.Errorf("hash = %q, want a bcrypt hash", hashedPassword)
	}
	if !auth.VerifyPassword(hashedPassword, plainPassword) {
		t.Error("VerifyPassword rejected the correct password")
	}
}

func TestHashPasswordUsesAFreshSalt(t *testing.T) {
	firstHash, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("first HashPassword: %v", err)
	}
	secondHash, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("second HashPassword: %v", err)
	}

	if firstHash == secondHash {
		t.Fatal("two hashes of the same password are identical: the salt is not random")
	}
	for _, hashedPassword := range []string{firstHash, secondHash} {
		if !auth.VerifyPassword(hashedPassword, plainPassword) {
			t.Errorf("VerifyPassword rejected %q", hashedPassword)
		}
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if auth.VerifyPassword(hashedPassword, plainPassword+"x") {
		t.Error("VerifyPassword accepted a wrong password")
	}
}

func TestVerifyPasswordRejectsEmptyOrBrokenHash(t *testing.T) {
	if auth.VerifyPassword("", plainPassword) {
		t.Error("VerifyPassword accepted an empty hash")
	}
	if auth.VerifyPassword("not-a-bcrypt-hash", plainPassword) {
		t.Error("VerifyPassword accepted a malformed hash")
	}
}

func TestHashPasswordRejectsEmptyPassword(t *testing.T) {
	if _, err := auth.HashPassword(""); !errors.Is(err, auth.ErrPasswordEmpty) {
		t.Fatalf("error = %v, want ErrPasswordEmpty", err)
	}
}

func TestHashPasswordRejectsTooLongPassword(t *testing.T) {
	// bcrypt 只处理前 72 字节；更长的密码必须报错，不能静默截断。
	if _, err := auth.HashPassword(strings.Repeat("a", 73)); !errors.Is(err, auth.ErrPasswordTooLong) {
		t.Fatalf("error = %v, want ErrPasswordTooLong", err)
	}
	if _, err := auth.HashPassword(strings.Repeat("a", 72)); err != nil {
		t.Fatalf("72 bytes should be accepted: %v", err)
	}
}
