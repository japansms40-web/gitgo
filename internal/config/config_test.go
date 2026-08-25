package config

import (
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"完整", Config{Token: "t", Owner: "o", Repo: "r", Branch: "main", Dir: "posts", IntervalSec: 1, Retries: 2}, false},
		{"缺Token", Config{Owner: "o", Repo: "r", Branch: "main"}, true},
		{"缺Owner", Config{Token: "t", Repo: "r", Branch: "main"}, true},
		{"缺Repo", Config{Token: "t", Owner: "o", Branch: "main"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cfg.Validate()
			if (err != nil) != c.wantErr {
				t.Fatalf("Validate() err=%v, wantErr=%v", err, c.wantErr)
			}
		})
	}
}

func TestNormalizeDefaults(t *testing.T) {
	cfg := Config{Token: "t", Owner: "o", Repo: "r"}
	cfg.Normalize()
	if cfg.Branch != "main" {
		t.Errorf("Branch 默认应为 main, got %q", cfg.Branch)
	}
	if cfg.IntervalSec < 0 || cfg.Retries < 0 {
		t.Errorf("间隔/重试不应为负")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	cfg := Config{
		Token: "t", Owner: "o", Repo: "r", Branch: "dev", Dir: "articles",
		AutoCreate: true, IntervalSec: 3, Retries: 5,
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if got := Load(path); got != cfg {
		t.Errorf("Load() = %+v, want %+v", got, cfg)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	want := Config{Branch: "main", Dir: "posts", IntervalSec: 1, Retries: 2}
	if got := Load(path); got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}
