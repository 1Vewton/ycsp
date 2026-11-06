package traceid_test

import (
	"context"
	"strings"
	"testing"

	"github.com/1Vewton/ycsp/backend/internal/traceid"
)

func TestNewUsesPrefixAndIsUnique(t *testing.T) {
	first := traceid.New("web")
	second := traceid.New("web")

	if !strings.HasPrefix(first, "web-") {
		t.Errorf("first id = %q, want a web- prefix", first)
	}
	if first == second {
		t.Errorf("two generated ids are identical: %q", first)
	}
	if len(first) > traceid.MaxLength {
		t.Errorf("id length = %d, want at most %d", len(first), traceid.MaxLength)
	}
}

func TestNormalizeAcceptsSafeIDs(t *testing.T) {
	for _, raw := range []string{"web-1234", " api-abc_def.1:2 ", "compute-9f8e"} {
		if got := traceid.Normalize(raw); got != strings.TrimSpace(raw) {
			t.Errorf("Normalize(%q) = %q, want the trimmed value", raw, got)
		}
	}
}

func TestNormalizeRejectsUnsafeIDs(t *testing.T) {
	longID := strings.Repeat("a", traceid.MaxLength+1)
	for _, raw := range []string{"", "   ", "web 123", "web-123\nforged", longID, "中文编号"} {
		if got := traceid.Normalize(raw); got != "" {
			t.Errorf("Normalize(%q) = %q, want an empty string", raw, got)
		}
	}
}

func TestContextRoundTrip(t *testing.T) {
	ctx := traceid.With(context.Background(), "web-123")

	if got := traceid.From(ctx); got != "web-123" {
		t.Errorf("From(ctx) = %q, want web-123", got)
	}
	if got := traceid.From(context.Background()); got != "" {
		t.Errorf("From(background) = %q, want empty", got)
	}
}
