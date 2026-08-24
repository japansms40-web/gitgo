package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestThemeColorUsesFixedVariant(t *testing.T) {
	light := NewTheme(theme.VariantLight)
	dark := NewTheme(theme.VariantDark)

	// 传入的 variant 参数应被忽略，颜色始终按 NewTheme 时固定的 Variant 取值。
	if got, want := light.Color(theme.ColorNameBackground, theme.VariantDark), color.Color(color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); got != want {
		t.Errorf("light.Color(Background) = %#v, want %#v", got, want)
	}
	if got, want := dark.Color(theme.ColorNameBackground, theme.VariantLight), color.Color(color.NRGBA{R: 0x16, G: 0x1B, B: 0x22, A: 0xFF}); got != want {
		t.Errorf("dark.Color(Background) = %#v, want %#v", got, want)
	}
}

func TestThemeColorFallsBackToDefaultTheme(t *testing.T) {
	// theme.DefaultTheme().Color 内部会读取 fyne.CurrentApp()，headless 测试
	// 环境下必须先用 test.NewApp() 起一个测试 app，否则会 panic。
	test.NewApp()
	light := NewTheme(theme.VariantLight)
	got := light.Color(theme.ColorNameFocus, theme.VariantLight)
	want := theme.DefaultTheme().Color(theme.ColorNameFocus, theme.VariantLight)
	if got != want {
		t.Errorf("fallback color = %#v, want %#v", got, want)
	}
}

func TestThemeAccessors(t *testing.T) {
	light := NewTheme(theme.VariantLight)
	cases := []struct {
		name string
		got  color.Color
		want color.Color
	}{
		{"accent", light.accentColor(), color.NRGBA{R: 0x1F, G: 0x6F, B: 0xEB, A: 0xFF}},
		{"success", light.successColor(), color.NRGBA{R: 0x1A, G: 0x7F, B: 0x37, A: 0xFF}},
		{"warning", light.warningColor(), color.NRGBA{R: 0x9A, G: 0x67, B: 0x00, A: 0xFF}},
		{"error", light.errorColor(), color.NRGBA{R: 0xCF, G: 0x22, B: 0x2E, A: 0xFF}},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %#v, want %#v", c.name, c.got, c.want)
		}
	}
}

func TestThemeSizeRadius(t *testing.T) {
	light := NewTheme(theme.VariantLight)
	if got := light.Size(theme.SizeNameCardRadius); got != 8 {
		t.Errorf("Size(CardRadius) = %v, want 8", got)
	}
	if got := light.Size(theme.SizeNameInputRadius); got != 6 {
		t.Errorf("Size(InputRadius) = %v, want 6", got)
	}
}

func TestBadgeTint(t *testing.T) {
	got := badgeTint(color.NRGBA{R: 0x1A, G: 0x7F, B: 0x37, A: 0xFF}, 36)
	want := color.NRGBA{R: 0x1A, G: 0x7F, B: 0x37, A: 36}
	if got != want {
		t.Errorf("badgeTint = %#v, want %#v", got, want)
	}
}
