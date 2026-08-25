package config

import "testing"

func TestNormalizeClampsKeywordSlots(t *testing.T) {
	cfg := Config{KeywordSlots: -5}
	cfg.Normalize()
	if cfg.KeywordSlots != 0 {
		t.Errorf("KeywordSlots = %d, want 0", cfg.KeywordSlots)
	}
}

func TestNormalizeKeepsPositiveKeywordSlots(t *testing.T) {
	cfg := Config{KeywordSlots: 3}
	cfg.Normalize()
	if cfg.KeywordSlots != 3 {
		t.Errorf("KeywordSlots = %d, want 3（合法值不应被改动）", cfg.KeywordSlots)
	}
}
