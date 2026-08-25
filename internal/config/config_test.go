package config

import (
	"path/filepath"
	"testing"
)

func TestNormalizeClampsInvalid(t *testing.T) {
	cfg := Config{
		Threads: -1, IntervalSec: -1, PerAccountCount: -1,
		FailSwitchCount: -1, CycleRounds: -1, RoundIntervalSec: -1, KeywordSlots: -1,
	}
	cfg.Normalize()
	want := Config{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	if cfg != want {
		t.Errorf("非法值应被纠正为最小合法值, got %+v, want %+v", cfg, want)
	}
}

func TestNormalizeKeepsPositive(t *testing.T) {
	cfg := Config{
		Threads: 10, IntervalSec: 2, PerAccountCount: 1000,
		FailSwitchCount: 100, CycleRounds: 3, RoundIntervalSec: 1, KeywordSlots: 3, CreateRepo: true,
	}
	want := cfg
	cfg.Normalize()
	if cfg != want {
		t.Errorf("合法值不应被改动, got %+v, want %+v", cfg, want)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	cfg := Config{
		Threads: 5, IntervalSec: 2, PerAccountCount: 50,
		FailSwitchCount: 10, CycleRounds: 2, RoundIntervalSec: 3, KeywordSlots: 2, CreateRepo: true,
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
	want := defaults()
	if got := Load(path); got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}
