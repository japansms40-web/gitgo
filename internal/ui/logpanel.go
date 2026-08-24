package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"githubbaidu/internal/publisher"
)

const logPanelMaxLines = 1000

// logTag 返回某类发布事件在日志面板中的标签文案与配色。
func logTag(t *Theme, kind publisher.EventKind) (tag string, tagColor, msgColor color.Color) {
	fg := t.foregroundColor()
	switch kind {
	case publisher.EventStart:
		return "[开始]", t.accentColor(), fg
	case publisher.EventSuccess:
		return "[成功]", t.successColor(), fg
	case publisher.EventFailure:
		return "[失败]", t.errorColor(), t.errorColor()
	case publisher.EventRetry:
		return "[重试]", t.warningColor(), t.warningColor()
	default:
		return "[信息]", t.accentColor(), fg
	}
}

// logLine 把一条发布事件格式化为日志正文（不含时间戳/标签）。
func logLine(e publisher.Event) string {
	switch e.Kind {
	case publisher.EventStart:
		return e.Title
	case publisher.EventSuccess:
		return fmt.Sprintf("%s → %s", e.Title, e.URL)
	case publisher.EventFailure:
		return fmt.Sprintf("%s 失败: %v", e.Title, e.Err)
	case publisher.EventRetry:
		return fmt.Sprintf("%s 重试: %v", e.Title, e.Err)
	default:
		return e.Title
	}
}

// LogPanel 是终端风格的运行日志面板：深底、等宽字体、按事件类型着色，
// 内部维护一份纯文本副本供复制/导出使用。历史行的颜色在写入时定格，
// 主题切换不会重新着色已有行（新写入的行会用新主题的颜色）。
type LogPanel struct {
	theme      *Theme
	rows       *fyne.Container
	scroll     *container.Scroll
	lines      []string
	autoScroll bool
}

// NewLogPanel 创建日志面板。
func NewLogPanel(t *Theme) *LogPanel {
	rows := container.NewVBox()
	scroll := container.NewVScroll(rows)
	return &LogPanel{theme: t, rows: rows, scroll: scroll, autoScroll: true}
}

// CanvasObject 返回可放入布局的滚动容器。
func (p *LogPanel) CanvasObject() fyne.CanvasObject { return p.scroll }

// SetAutoScroll 开关"新日志自动滚动到底部"。
func (p *LogPanel) SetAutoScroll(on bool) { p.autoScroll = on }

// AppendEvent 追加一条发布事件日志。
func (p *LogPanel) AppendEvent(e publisher.Event) {
	tag, tagColor, msgColor := logTag(p.theme, e.Kind)
	p.append(tag, tagColor, logLine(e), msgColor)
}

// AppendInfo 追加一条普通信息日志（非发布事件，如启动/完成提示）。
func (p *LogPanel) AppendInfo(msg string) {
	p.append("[信息]", p.theme.accentColor(), msg, p.theme.foregroundColor())
}

func (p *LogPanel) append(tag string, tagColor color.Color, msg string, msgColor color.Color) {
	ts := time.Now().Format("15:04:05")
	p.lines = append(p.lines, fmt.Sprintf("%s %s %s", ts, tag, msg))
	if len(p.lines) > logPanelMaxLines {
		p.lines = p.lines[len(p.lines)-logPanelMaxLines:]
		p.rows.Objects = p.rows.Objects[1:]
	}

	tsText := canvas.NewText(ts, p.theme.mutedColor())
	tsText.TextStyle = fyne.TextStyle{Monospace: true}
	tsText.TextSize = 12
	tagText := canvas.NewText(tag, tagColor)
	tagText.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	tagText.TextSize = 12
	msgText := canvas.NewText(msg, msgColor)
	msgText.TextStyle = fyne.TextStyle{Monospace: true}
	msgText.TextSize = 12

	p.rows.Add(container.NewHBox(tsText, tagText, msgText))
	if p.autoScroll {
		p.scroll.ScrollToBottom()
	}
}

// Clear 清空日志。
func (p *LogPanel) Clear() {
	p.lines = nil
	p.rows.Objects = nil
	p.rows.Refresh()
}

// Text 返回全部日志的纯文本形式，用于复制/导出。
func (p *LogPanel) Text() string { return strings.Join(p.lines, "\n") }

// Count 返回当前日志行数。
func (p *LogPanel) Count() int { return len(p.lines) }
