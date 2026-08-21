package config

import "testing"

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
