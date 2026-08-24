package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// badgeStyle 把发布状态文本映射为状态徽标的背景色、文字色。
// "失败: xxx" 这类带详情后缀的状态按前缀 "失败" 归类为失败态。
func badgeStyle(t *Theme, state string) (bg, fg color.Color) {
	switch {
	case state == "成功":
		c := t.successColor()
		return badgeTint(c, 36), c
	case strings.HasPrefix(state, "失败"):
		c := t.errorColor()
		return badgeTint(c, 36), c
	case state == "发布中" || strings.HasPrefix(state, "重试中"):
		c := t.warningColor()
		return badgeTint(c, 40), c
	default: // 待发布
		return color.Transparent, t.mutedColor()
	}
}

// badge 是可复用的圆角状态徽标 CanvasObject，配合 widget.Table 的
// CreateCell/UpdateCell 复用模式：同一个 *badge 会在不同行之间被反复 set()。
// object() 始终返回同一个已构建好的容器实例，这样调用方在其上调用
// Hide()/Show() 才能真正影响表格里显示的那个对象（而不是每次新建一份）。
type badge struct {
	bg   *canvas.Rectangle
	txt  *canvas.Text
	root fyne.CanvasObject
}

func newBadge() *badge {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 8
	txt := canvas.NewText("", color.Black)
	txt.TextSize = 12
	txt.Alignment = fyne.TextAlignCenter
	pill := container.NewStack(bg, container.NewPadded(txt))
	root := container.NewHBox(pill, layout.NewSpacer())
	return &badge{bg: bg, txt: txt, root: root}
}

func (b *badge) object() fyne.CanvasObject { return b.root }

// set 更新徽标的颜色与文字并触发重绘。非并发安全，
// 必须只在 Fyne 主线程（fyne.Do 内）调用。
func (b *badge) set(t *Theme, state string) {
	bg, fg := badgeStyle(t, state)
	b.bg.FillColor = bg
	b.txt.Color = fg
	b.txt.Text = state
	b.bg.Refresh()
	b.txt.Refresh()
}
