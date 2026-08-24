package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/theme"
)

func TestBadgeStyleSuccess(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	bg, fg := badgeStyle(th, "成功")
	if fg != th.successColor() {
		t.Errorf("fg = %#v, want success color", fg)
	}
	nrgba, ok := bg.(color.NRGBA)
	if !ok || nrgba.A != 36 {
		t.Errorf("bg = %#v, want NRGBA alpha 36", bg)
	}
}

func TestBadgeStyleFailureWithDetail(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	// "失败: xxx" 这类带详情后缀的状态也要按失败态显示。
	_, fg := badgeStyle(th, "失败: boom")
	if fg != th.errorColor() {
		t.Errorf("fg = %#v, want error color", fg)
	}
}

func TestBadgeStyleRunningAndRetry(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	_, fgRunning := badgeStyle(th, "发布中")
	_, fgRetry := badgeStyle(th, "重试中: net")
	if fgRunning != th.warningColor() || fgRetry != th.warningColor() {
		t.Errorf("running/retry fg = %#v / %#v, want warning color", fgRunning, fgRetry)
	}
}

func TestBadgeStylePending(t *testing.T) {
	th := NewTheme(theme.VariantLight)
	bg, fg := badgeStyle(th, "待发布")
	if bg != color.Transparent {
		t.Errorf("bg = %#v, want transparent", bg)
	}
	if fg != th.mutedColor() {
		t.Errorf("fg = %#v, want muted color", fg)
	}
}
