package checkpoint

import (
	"context"
	"testing"
)

func TestSafeCommand(t *testing.T) {
	ctx := context.Background()

	t.Run("Allowed tool", func(t *testing.T) {
		cmd, err := safeCommand(ctx, "go", "version")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}
	})

	t.Run("Disallowed tool", func(t *testing.T) {
		_, err := safeCommand(ctx, "ls")
		if err == nil {
			t.Error("expected error for disallowed tool")
		}
	})

	t.Run("Invalid argument", func(t *testing.T) {
		_, err := safeCommand(ctx, "go", "test; rm -rf /")
		if err == nil {
			t.Error("expected error for invalid argument")
		}
	})

	t.Run("Tool not found", func(t *testing.T) {
		_, err := safeCommand(ctx, "gosec", "-version")
		// gosec might not be installed, so we expect either success or "failed to find tool"
		if err != nil && !testing.Short() {
			t.Logf("gosec not found as expected: %v", err)
		}
	})
}
