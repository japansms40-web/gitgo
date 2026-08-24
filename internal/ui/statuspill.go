package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// pillText 根据发布状态生成标题栏状态胶囊文案。
func pillText(running bool, done, total, failed int) string {
	switch {
	case running:
		return fmt.Sprintf("发布中 · %d/%d", done, total)
	case total == 0 || done == 0:
		return "待机"
	case failed > 0:
		return fmt.Sprintf("已完成 · %d 成功 · %d 失败", done-failed, failed)
	default:
		return fmt.Sprintf("已完成 · %d/%d", done, total)
	}
}

// pillKind 标识状态胶囊的语义色。
type pillKind int

const (
	pillMuted pillKind = iota
	pillRunning
	pillSuccess
	pillError
)

func pillStateKind(running bool, done, total, failed int) pillKind {
	switch {
	case running:
		return pillRunning
	case total == 0 || done == 0:
		return pillMuted
	case failed > 0:
		return pillError
	default:
		return pillSuccess
	}
}

// formatElapsed 把耗时格式化为 mm:ss（超过一小时则 hh:mm:ss）。
func formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// pillStyle 返回状态胶囊在给定状态下的背景色、文字色。文字色直接来自
// 主题取色，不经过 widget.Label 的 Importance 机制——后者的 LowImportance
// 会映射到 Fyne 默认主题里未被本项目定制的 disabled 色，与胶囊自身的浅色
// 背景几乎融为一体、肉眼不可见。
func pillStyle(t *Theme, running bool, done, total, failed int) (bg, fg color.Color) {
	switch pillStateKind(running, done, total, failed) {
	case pillRunning:
		fg = t.accentColor()
	case pillSuccess:
		fg = t.successColor()
	case pillError:
		fg = t.errorColor()
	default:
		fg = t.mutedColor()
	}
	return badgeTint(fg, 36), fg
}

// statusPill 是标题栏右侧的圆角状态胶囊。running/done/total/failed 变化后
// （每条发布事件、主题切换后）都需要调用一次 repaint 来刷新文案与颜色。
type statusPill struct {
	bg  *canvas.Rectangle
	txt *canvas.Text
}

func newStatusPill() *statusPill {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 12
	txt := canvas.NewText("待机", color.Black)
	return &statusPill{bg: bg, txt: txt}
}

func (p *statusPill) object() fyne.CanvasObject {
	return container.NewStack(p.bg, container.NewPadded(p.txt))
}

// repaint 根据当前状态更新胶囊文案与颜色并触发重绘。非并发安全，
// 必须只在 Fyne 主线程（fyne.Do 内）调用。
func (p *statusPill) repaint(running bool, done, total, failed int) {
	bg, fg := pillStyle(currentTheme(), running, done, total, failed)
	p.bg.FillColor = bg
	p.txt.Color = fg
	p.txt.Text = pillText(running, done, total, failed)
	p.bg.Refresh()
	p.txt.Refresh()
}
