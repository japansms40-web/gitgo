package account

import (
	"path/filepath"
	"testing"
)

func TestParseImportText(t *testing.T) {
	text := "  ck-one \n----\nck-two\n----\n   \n----\nck-three"
	got := ParseImportText(text)
	want := []string{"ck-one", "ck-two", "ck-three"}
	if len(got) != len(want) {
		t.Fatalf("解析出 %d 个账号, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].CK != w {
			t.Errorf("got[%d].CK = %q, want %q", i, got[i].CK, w)
		}
		if got[i].Status != StatusPending {
			t.Errorf("got[%d].Status = %q, want %q", i, got[i].Status, StatusPending)
		}
	}
}

func TestParseImportTextEmpty(t *testing.T) {
	if got := ParseImportText("   \n----\n\n"); len(got) != 0 {
		t.Errorf("空文本应解析出 0 个账号, got %+v", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "accounts.json")
	list := []Account{
		{CK: "a", UA: "Chrome", IP: "1.2.3.4", Status: StatusSuccess, Success: 3, Total: 3},
		{CK: "b", Status: StatusFailed, Fail: 1, Total: 1, Bad: true},
	}
	if err := Save(path, list); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got := Load(path)
	if len(got) != len(list) {
		t.Fatalf("Load() = %+v, want %+v", got, list)
	}
	for i := range list {
		if got[i] != list[i] {
			t.Errorf("Load()[%d] = %+v, want %+v", i, got[i], list[i])
		}
	}
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	if got := Load(path); len(got) != 0 {
		t.Errorf("Load() = %+v, want empty", got)
	}
}
