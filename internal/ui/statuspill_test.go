package ui

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2/theme"
)

func TestPillTextIdle(t *testing.T) {
	if got := pillText(false, 0, 0, 0); got != "待机" {
		t.Errorf("pillText idle = %q", got)
	}
}

func TestPillTextRunning(t *testing.T) {
	if got := pillText(true, 3, 10, 0); got != "发布中 · 3/10" {
		t.Errorf("pillText running = %q", got)
	}
}

func TestPillTextDoneAllSuccess(t *testing.T) {
	if got := pillText(false, 10, 10, 0); got != "已完成 · 10/10" {
		t.Errorf("pillText done = %q", got)
	}
}

func TestPillTextDoneWithFailures(t *testing.T) {
	if got := pillText(false, 10, 10, 3); got != "已完成 · 7 成功 · 3 失败" {
		t.Errorf("pillText done with failures = %q", got)
	}
}

func TestPillStateKind(t *testing.T) {
	cases := []struct {
		name                string
		running             bool
		done, total, failed int
		want                pillKind
	}{
		{"idle", false, 0, 0, 0, pillMuted},
		{"running", true, 2, 5, 0, pillRunning},
		{"success", false, 5, 5, 0, pillSuccess},
		{"error", false, 5, 5, 2, pillError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := pillStateKind(c.running, c.done, c.total, c.failed); got != c.want {
				t.Errorf("pillStateKind(%v,%d,%d,%d) = %v, want %v", c.running, c.done, c.total, c.failed, got, c.want)
			}
		})
	}
}

// TestPillStyleIdleUsesThemedMutedColor 是回归测试：待机态文字曾经通过
// widget.LowImportance 间接取色，落到 Fyne 默认主题里近乎纯白的
// disabled 色（#E3E3E3），与胶囊自身的浅色背景几乎融为一体、肉眼不可见。
// 这里要求 fg 直接等于本主题的 mutedColor()，确保与胶囊底色有足够对比度。
func TestPillStyleIdleUsesThemedMutedColor(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	bg, fg := pillStyle(th, false, 0, 0, 0)
	if fg != th.mutedColor() {
		t.Errorf("fg = %#v, want muted color %#v", fg, th.mutedColor())
	}
	nrgba, ok := bg.(color.NRGBA)
	if !ok || nrgba.A != 36 {
		t.Errorf("bg = %#v, want NRGBA alpha 36", bg)
	}
}

func TestPillStyleRunning(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	_, fg := pillStyle(th, true, 2, 5, 0)
	if fg != th.accentColor() {
		t.Errorf("fg = %#v, want accent color", fg)
	}
}

func TestPillStyleSuccess(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	_, fg := pillStyle(th, false, 5, 5, 0)
	if fg != th.successColor() {
		t.Errorf("fg = %#v, want success color", fg)
	}
}

func TestPillStyleError(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	_, fg := pillStyle(th, false, 5, 5, 2)
	if fg != th.errorColor() {
		t.Errorf("fg = %#v, want error color", fg)
	}
}

func TestFormatElapsed(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00"},
		{27 * time.Second, "00:27"},
		{90 * time.Second, "01:30"},
		{61*time.Minute + 5*time.Second, "01:01:05"},
	}
	for _, c := range cases {
		if got := formatElapsed(c.d); got != c.want {
			t.Errorf("formatElapsed(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}
