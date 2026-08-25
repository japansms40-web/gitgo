package account

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPathEndsWithAccountsJSON(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if filepath.Base(path) != "accounts.json" {
		t.Errorf("DefaultPath() = %q, want basename accounts.json", path)
	}
	if !strings.Contains(path, appName) {
		t.Errorf("DefaultPath() = %q, want it to contain app dir %q", path, appName)
	}
}
