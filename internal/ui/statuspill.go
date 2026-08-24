package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

// statusPill 是标题栏右侧的圆角状态胶囊。running/done/total/failed 变化后
// （每条发布事件、主题切换后）都需要调用一次 repaint 来刷新文案与颜色。
type statusPill struct {
	bg  *canvas.Rectangle
	lbl *widget.Label
}

func newStatusPill() *statusPill {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 12
	lbl := widget.NewLabel("待机")
	return &statusPill{bg: bg, lbl: lbl}
}

func (p *statusPill) object() fyne.CanvasObject {
	return container.NewStack(p.bg, container.NewPadded(p.lbl))
}

func (p *statusPill) repaint(running bool, done, total, failed int) {
	t := currentTheme()
	var fg color.Color
	switch pillStateKind(running, done, total, failed) {
	case pillRunning:
		fg = t.accentColor()
		p.lbl.Importance = widget.HighImportance
	case pillSuccess:
		fg = t.successColor()
		p.lbl.Importance = widget.SuccessImportance
	case pillError:
		fg = t.errorColor()
		p.lbl.Importance = widget.DangerImportance
	default:
		fg = t.mutedColor()
		p.lbl.Importance = widget.LowImportance
	}
	p.bg.FillColor = badgeTint(fg, 36)
	p.lbl.SetText(pillText(running, done, total, failed))
	p.bg.Refresh()
}
