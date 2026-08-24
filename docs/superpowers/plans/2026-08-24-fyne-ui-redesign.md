# Fyne UI 视觉重设计 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `docs/superpowers/specs/2026-08-24-fyne-ui-redesign-design.md` 里确定的视觉风格（配色、圆角卡片、终端日志面板、状态徽标、深浅主题）套到现有的单账号 GitHub Markdown 发布器上，业务逻辑不变。

**Architecture:** 新增几个职责单一的小文件（自定义 `fyne.Theme`、状态徽标、状态胶囊、终端日志面板、帮助页），每个文件里把"可以脱离 Fyne 运行时单元测试的纯函数"和"需要手动运行验证的 Fyne 组件装配代码"分开写，再在重写后的 `mainwindow.go` 里把它们组装成新外壳（标题栏 + 左侧导航 + 发布页/帮助页 + 右侧参数栏 + 日志面板 + 状态栏）。

**Tech Stack:** Go 1.26、Fyne v2.8.0（`fyne.io/fyne/v2`，已在 `go.mod` 声明，无需新增依赖）。

---

## 开工前须知

- 本仓库当前没有用 git worktree 隔离开发；直接在 `main` 分支上按任务提交即可（和现有两次提交的习惯一致）。
- 已用以下方式确认过本任务用到的所有 Fyne API 在 v2.8.0 中真实存在且签名如下（避免计划里出现编译不过的代码）：
  - `fyne.Theme` 接口：`Color(ThemeColorName, ThemeVariant) color.Color`、`Font(TextStyle) Resource`、`Icon(ThemeIconName) Resource`、`Size(ThemeSizeName) float32`
  - `theme.VariantLight` / `theme.VariantDark`、`theme.DefaultTheme() fyne.Theme`
  - `theme.ColorNameBackground/MenuBackground/InputBackground/Foreground/PlaceHolder/InputBorder/Separator/Primary/Selection/Success/Warning/Error`
  - `theme.SizeNameInputRadius/SelectionRadius/ButtonRadius/CardRadius`
  - `canvas.Rectangle.CornerRadius float32`、`canvas.Text{Color, TextSize, TextStyle, Alignment}`
  - `container.NewStack/NewCenter/NewPadded/NewHBox/NewVBox/NewBorder/NewVScroll/NewGridWithColumns`、`container.New(layout.Layout, ...)`
  - `layout.NewGridWrapLayout(fyne.Size) fyne.Layout`、`layout.NewSpacer() fyne.CanvasObject`
  - `widget.Table` 的 `CreateCell`/`UpdateCell` 回调复用同一个对象池：`UpdateCell` 收到的 `fyne.CanvasObject` 与对应 `CreateCell()` 返回的是同一个指针（已读 Fyne 源码 `widget/table.go` + `widget/list.go` 的 `createItemAndApplyThemeScope` 确认，未被包装）。
  - `widget.Label.Importance`（`LowImportance/MediumImportance/HighImportance/DangerImportance/WarningImportance/SuccessImportance`）会驱动文字颜色，且随主题变化自动刷新。
  - `widget.Button.Disabled() bool`（不是 `Enabled()`）。
  - `container.Scroll.ScrollToBottom()`。
  - `fyne.App.Clipboard() Clipboard`（自 2.6 起），`Clipboard.SetContent(string)`。
  - `fyne.Settings.AddListener(func(fyne.Settings))`（自 2.6 起，"guaranteed to be invoked on the app goroutine"）。
- **已实测踩过的坑，计划里的代码已经按这个来写，照抄即可，不要再"优化"掉：**
  - `theme.DefaultTheme().Color(...)`（我们 `Theme.Color` 的兜底分支）内部会读 `fyne.CurrentApp()`，在没有启动 Fyne app 的 `go test` 环境下直接 panic。凡是测试会走到这个兜底分支的用例，必须先调用 `fyne.io/fyne/v2/test` 包的 `test.NewApp()`（Task 1 的 `TestThemeColorFallsBackToDefaultTheme` 已经这么写）。
  - `canvas.Text`/`canvas.Rectangle` 参与布局时（比如放进 `container.NewHBox` 再 `.Add()`）会触发文字尺寸测量，同样需要 `fyne.CurrentApp()`；即使用 `test.NewApp()` 起了测试 app，Fyne 内置的测试主题也只注册了少数几种字体样式组合（`{Monospace:true}`、`{Bold:true, Italic:true}` 等），我们日志行用的 `{Monospace:true, Bold:true}` 组合没有注册，一样会 panic。所以 `LogPanel`（`logpanel.go`）的 `append`/`AppendEvent`/`AppendInfo`/`Clear`/`Text`/`Count` 这类会真正构建/摆放 canvas 对象的方法**不写单测**，只测 `logTag`/`logLine` 这两个不碰 canvas 的纯函数（Task 4 已经是这样），`LogPanel` 本身的行为放到 Task 6 Step 4 的手动验证清单里过一遍。
- 现有 `internal/ui/state.go` / `state_test.go` 完全不改，`RowState` 的公开方法（`Status`/`URL`/`Done`/`Progress`/`SuccessURLs`）保持原样，新代码只是调用方。

---

### Task 1: 自定义主题配色 `theme.go`

**Files:**
- Create: `internal/ui/theme.go`
- Test: `internal/ui/theme_test.go`

- [ ] **Step 1: 写失败的测试**

创建 `internal/ui/theme_test.go`：

```go
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/ui/... -run 'TestTheme|TestBadgeTint' -v`
Expected: FAIL（`NewTheme`/`badgeTint` 等未定义，编译错误）

- [ ] **Step 3: 实现 `theme.go`**

创建 `internal/ui/theme.go`：

```go
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Theme 套用设计稿配色的自定义主题。Variant 在创建时固定，不随 Fyne 传入的
// variant 参数变化——用于支持应用内手动切换 Light/Dark，与 OS 主题设置解耦。
type Theme struct {
	Variant fyne.ThemeVariant
}

// NewTheme 返回固定为 variant 的主题实例。
func NewTheme(variant fyne.ThemeVariant) *Theme {
	return &Theme{Variant: variant}
}

var lightPalette = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
	theme.ColorNameMenuBackground:  color.NRGBA{R: 0xF3, G: 0xF5, B: 0xF7, A: 0xFF},
	theme.ColorNameInputBackground: color.NRGBA{R: 0xF6, G: 0xF7, B: 0xF9, A: 0xFF},
	theme.ColorNameForeground:      color.NRGBA{R: 0x1A, G: 0x1E, B: 0x23, A: 0xFF},
	theme.ColorNamePlaceHolder:     color.NRGBA{R: 0x6E, G: 0x76, B: 0x81, A: 0xFF},
	theme.ColorNameInputBorder:     color.NRGBA{R: 0xB6, G: 0xBD, B: 0xC7, A: 0xFF},
	theme.ColorNameSeparator:       color.NRGBA{R: 0xDC, G: 0xE0, B: 0xE6, A: 0xFF},
	theme.ColorNamePrimary:         color.NRGBA{R: 0x1F, G: 0x6F, B: 0xEB, A: 0xFF},
	theme.ColorNameSelection:       color.NRGBA{R: 0xE7, G: 0xF0, B: 0xFE, A: 0xFF},
	theme.ColorNameSuccess:         color.NRGBA{R: 0x1A, G: 0x7F, B: 0x37, A: 0xFF},
	theme.ColorNameWarning:         color.NRGBA{R: 0x9A, G: 0x67, B: 0x00, A: 0xFF},
	theme.ColorNameError:           color.NRGBA{R: 0xCF, G: 0x22, B: 0x2E, A: 0xFF},
}

var darkPalette = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.NRGBA{R: 0x16, G: 0x1B, B: 0x22, A: 0xFF},
	theme.ColorNameMenuBackground:  color.NRGBA{R: 0x11, G: 0x16, B: 0x1D, A: 0xFF},
	theme.ColorNameInputBackground: color.NRGBA{R: 0x1B, G: 0x21, B: 0x29, A: 0xFF},
	theme.ColorNameForeground:      color.NRGBA{R: 0xE6, G: 0xED, B: 0xF3, A: 0xFF},
	theme.ColorNamePlaceHolder:     color.NRGBA{R: 0x8B, G: 0x94, B: 0x9E, A: 0xFF},
	theme.ColorNameInputBorder:     color.NRGBA{R: 0x3D, G: 0x44, B: 0x4D, A: 0xFF},
	theme.ColorNameSeparator:       color.NRGBA{R: 0x2A, G: 0x31, B: 0x3A, A: 0xFF},
	theme.ColorNamePrimary:         color.NRGBA{R: 0x38, G: 0x8B, B: 0xFD, A: 0xFF},
	theme.ColorNameSelection:       color.NRGBA{R: 0x13, G: 0x2A, B: 0x4D, A: 0xFF},
	theme.ColorNameSuccess:         color.NRGBA{R: 0x3F, G: 0xB9, B: 0x50, A: 0xFF},
	theme.ColorNameWarning:         color.NRGBA{R: 0xD2, G: 0x99, B: 0x22, A: 0xFF},
	theme.ColorNameError:           color.NRGBA{R: 0xF8, G: 0x51, B: 0x49, A: 0xFF},
}

// LogBackground 是底部终端日志面板的固定深底色，Light/Dark 主题下一致。
var LogBackground = color.NRGBA{R: 0x0B, G: 0x0E, B: 0x12, A: 0xFF}

func (t *Theme) palette() map[fyne.ThemeColorName]color.Color {
	if t.Variant == theme.VariantDark {
		return darkPalette
	}
	return lightPalette
}

// Color 实现 fyne.Theme。传入的 variant 参数被忽略，颜色始终按 t.Variant 取值，
// 以支持应用内手动切换主题（与 Fyne 依据 OS 传入的 variant 解耦）。
func (t *Theme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if c, ok := t.palette()[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, t.Variant)
}

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius, theme.SizeNameButtonRadius:
		return 6
	case theme.SizeNameCardRadius:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func (t *Theme) accentColor() color.Color     { return t.Color(theme.ColorNamePrimary, t.Variant) }
func (t *Theme) successColor() color.Color    { return t.Color(theme.ColorNameSuccess, t.Variant) }
func (t *Theme) warningColor() color.Color    { return t.Color(theme.ColorNameWarning, t.Variant) }
func (t *Theme) errorColor() color.Color      { return t.Color(theme.ColorNameError, t.Variant) }
func (t *Theme) mutedColor() color.Color      { return t.Color(theme.ColorNamePlaceHolder, t.Variant) }
func (t *Theme) foregroundColor() color.Color { return t.Color(theme.ColorNameForeground, t.Variant) }

// badgeTint 返回某语义色在给定透明度(0-255)下的低饱和版本，用于徽标/胶囊底色。
func badgeTint(base color.Color, alpha uint8) color.Color {
	r, g, b, _ := base.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: alpha}
}

// currentTheme 返回当前应用生效的 *Theme。要求应用启动时已调用过
// app.Settings().SetTheme(NewTheme(...))（由 mainwindow.go 的 initialTheme 完成）。
func currentTheme() *Theme {
	return fyne.CurrentApp().Settings().Theme().(*Theme)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/ui/... -run 'TestTheme|TestBadgeTint' -v`
Expected: PASS（全部 5 个用例）

- [ ] **Step 5: 提交**

```bash
git add internal/ui/theme.go internal/ui/theme_test.go
git commit -m "$(cat <<'EOF'
feat(ui): 添加套用设计稿配色的自定义 Fyne 主题

实现 fyne.Theme 接口，Light/Dark 两套配色严格取自设计稿的
Fyne ColorName 对照表；Variant 固定不随传入参数变化，
用于支持应用内手动切换主题。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 2: 状态徽标 `badge.go`

**Files:**
- Create: `internal/ui/badge.go`
- Test: `internal/ui/badge_test.go`

- [ ] **Step 1: 写失败的测试**

创建 `internal/ui/badge_test.go`：

```go
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/ui/... -run TestBadgeStyle -v`
Expected: FAIL（`badgeStyle` 未定义）

- [ ] **Step 3: 实现 `badge.go`**

创建 `internal/ui/badge.go`：

```go
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

func (b *badge) set(t *Theme, state string) {
	bg, fg := badgeStyle(t, state)
	b.bg.FillColor = bg
	b.txt.Color = fg
	b.txt.Text = state
	b.bg.Refresh()
	b.txt.Refresh()
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/ui/... -run TestBadgeStyle -v`
Expected: PASS（4 个用例）

- [ ] **Step 5: 编译检查**

Run: `go build ./...`
Expected: 无输出，退出码 0

- [ ] **Step 6: 提交**

```bash
git add internal/ui/badge.go internal/ui/badge_test.go
git commit -m "$(cat <<'EOF'
feat(ui): 添加状态徽标组件

badgeStyle 把发布状态文本映射为语义色；badge 是配合
widget.Table CreateCell/UpdateCell 复用模式的圆角徽标对象。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 3: 状态胶囊与耗时格式化 `statuspill.go`

**Files:**
- Create: `internal/ui/statuspill.go`
- Test: `internal/ui/statuspill_test.go`

- [ ] **Step 1: 写失败的测试**

创建 `internal/ui/statuspill_test.go`：

```go
package ui

import (
	"testing"
	"time"
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/ui/... -run 'TestPill|TestFormatElapsed' -v`
Expected: FAIL（`pillText`/`pillStateKind`/`formatElapsed`/`pillKind` 未定义）

- [ ] **Step 3: 实现 `statuspill.go`**

创建 `internal/ui/statuspill.go`：

```go
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
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/ui/... -run 'TestPill|TestFormatElapsed' -v`
Expected: PASS（全部用例）

- [ ] **Step 5: 编译检查**

Run: `go build ./...`
Expected: 无输出，退出码 0

- [ ] **Step 6: 提交**

```bash
git add internal/ui/statuspill.go internal/ui/statuspill_test.go
git commit -m "$(cat <<'EOF'
feat(ui): 添加标题栏状态胶囊与耗时格式化

pillText/pillStateKind 根据运行状态与完成/失败计数生成
胶囊文案与语义色；formatElapsed 格式化发布耗时。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 4: 终端风格日志面板 `logpanel.go`

**Files:**
- Create: `internal/ui/logpanel.go`
- Test: `internal/ui/logpanel_test.go`

- [ ] **Step 1: 写失败的测试**

创建 `internal/ui/logpanel_test.go`。`LogPanel` 里真正构建/摆放
`canvas.Text`/`canvas.Rectangle` 的方法（`append`/`AppendEvent`/`AppendInfo`/
`Clear`/`Text`/`Count`）不写单测——实测过，即使用 `fyne.io/fyne/v2/test` 的
`test.NewApp()` 起一个 headless 测试 app，Fyne 内置测试主题也没有注册日志行
用到的 `{Monospace:true, Bold:true}` 字体组合，一样会 panic。这部分行为放到
Task 6 Step 4 的手动验证清单里过一遍。这里只测完全不碰 canvas 对象的两个纯
函数：

```go
package ui

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2/theme"

	"githubbaidu/internal/publisher"
)

func TestLogTagColors(t *testing.T) {
	th := NewTheme(theme.VariantLight)

	tag, tagColor, _ := logTag(th, publisher.EventSuccess)
	if tag != "[成功]" || tagColor != th.successColor() {
		t.Errorf("success tag = %q/%#v", tag, tagColor)
	}

	tag, tagColor, msgColor := logTag(th, publisher.EventFailure)
	if tag != "[失败]" || tagColor != th.errorColor() || msgColor != th.errorColor() {
		t.Errorf("failure tag = %q/%#v/%#v", tag, tagColor, msgColor)
	}

	tag, tagColor, _ = logTag(th, publisher.EventRetry)
	if tag != "[重试]" || tagColor != th.warningColor() {
		t.Errorf("retry tag = %q/%#v", tag, tagColor)
	}

	tag, tagColor, _ = logTag(th, publisher.EventStart)
	if tag != "[开始]" || tagColor != th.accentColor() {
		t.Errorf("start tag = %q/%#v", tag, tagColor)
	}
}

func TestLogLineFormatting(t *testing.T) {
	cases := []struct {
		event publisher.Event
		want  string
	}{
		{publisher.Event{Kind: publisher.EventStart, Title: "hello"}, "hello"},
		{publisher.Event{Kind: publisher.EventSuccess, Title: "hello", URL: "http://x/y"}, "hello → http://x/y"},
		{publisher.Event{Kind: publisher.EventFailure, Title: "hello", Err: errors.New("boom")}, "hello 失败: boom"},
		{publisher.Event{Kind: publisher.EventRetry, Title: "hello", Err: errors.New("net")}, "hello 重试: net"},
	}
	for _, c := range cases {
		if got := logLine(c.event); got != c.want {
			t.Errorf("logLine(%+v) = %q, want %q", c.event, got, c.want)
		}
	}
}

```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/ui/... -run 'TestLog' -v`
Expected: FAIL（`logTag`/`logLine` 未定义）

- [ ] **Step 3: 实现 `logpanel.go`**

创建 `internal/ui/logpanel.go`：

```go
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
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/ui/... -run 'TestLog' -v`
Expected: PASS（全部用例）

- [ ] **Step 5: 编译检查**

Run: `go build ./...`
Expected: 无输出，退出码 0

- [ ] **Step 6: 提交**

```bash
git add internal/ui/logpanel.go internal/ui/logpanel_test.go
git commit -m "$(cat <<'EOF'
feat(ui): 添加终端风格运行日志面板

logTag/logLine 按发布事件类型生成标签与配色；LogPanel 维护
可视化日志行 + 纯文本副本，供复制/导出/清空使用。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 5: 帮助页 `help.go`

**Files:**
- Create: `internal/ui/help.go`

- [ ] **Step 1: 实现 `help.go`**

内容取自现有 `README.md` 的"使用"与"说明"部分，创建 `internal/ui/help.go`：

```go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type helpItem struct {
	label string
	body  string
}

var helpBasics = []helpItem{
	{"权限", "在 GitHub 生成 Personal Access Token，需勾选 repo 权限；若要自动建仓库还需 public_repo。"},
	{"格式", "支持导入 .md / .txt 文件，.txt 会以 .md 提交，文件名作为标题。"},
	{"发布", "单账号、单仓库；文件已存在则更新覆盖，每篇一次提交。"},
	{"重试", "可设置发布间隔与失败重试次数；遇 API 限流会按 Retry-After 等待重试。"},
	{"配置", "设置（含 Token）保存在系统的 Fyne Preferences 中，下次打开自动载入。"},
}

// BuildHelp 构建"帮助"页内容：静态的使用说明卡片列表，内容取自 README。
func BuildHelp() fyne.CanvasObject {
	rows := container.NewVBox()
	for _, item := range helpBasics {
		body := widget.NewLabel(item.body)
		body.Wrapping = fyne.TextWrapWord
		rows.Add(widget.NewCard(item.label, "", body))
	}
	return container.NewVScroll(rows)
}
```

- [ ] **Step 2: 编译检查**

Run: `go build ./...`
Expected: 无输出，退出码 0

- [ ] **Step 3: 提交**

```bash
git add internal/ui/help.go
git commit -m "$(cat <<'EOF'
feat(ui): 添加帮助页，展示 README 使用说明

静态卡片列表，内容摘自现有 README（权限/格式/发布行为/
重试/配置保存位置），不含设计稿里的外部工具链接。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 6: 重写主窗口外壳 `mainwindow.go`

**Files:**
- Modify: `internal/ui/mainwindow.go`（全量重写）

这一步把 Task 1–5 的组件组装成设计文档里的外壳：标题栏（图标+名称+状态胶囊+主题切换）+ 左侧导航（发布/帮助）+ 发布页（工具条+设置卡片+队列表）/帮助页 + 右侧参数栏（只读摘要+开始/停止+保存/清空/导出）+ 底部日志面板 + 底部状态栏。业务逻辑（读配置、扫描队列、发布、导出链接）与原文件保持一致，只是重新接线。

- [ ] **Step 1: 用以下内容整体替换 `internal/ui/mainwindow.go`**

```go
package ui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
	"githubbaidu/internal/github"
	"githubbaidu/internal/publisher"
)

const keyThemeVariant = "themeVariant"

// initialTheme 根据已保存的偏好或系统设置决定启动时使用的主题变体，
// 并把它设为当前应用主题（后续所有 currentTheme() 调用据此取色）。
func initialTheme(prefs fyne.Preferences) *Theme {
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	switch prefs.String(keyThemeVariant) {
	case "light":
		variant = theme.VariantLight
	case "dark":
		variant = theme.VariantDark
	}
	t := NewTheme(variant)
	fyne.CurrentApp().Settings().SetTheme(t)
	return t
}

// themedRect 是随主题切换需要重新取色的结构背景矩形。
type themedRect struct {
	rect *canvas.Rectangle
	name fyne.ThemeColorName
}

// tableCell 是发布队列表格里一个单元格的可复用对象：文件名/仓库路径列显示
// label，状态列显示 badge，同一时刻只显示其中一个。
type tableCell struct {
	label *widget.Label
	badge *badge
}

// Build 装配并返回主窗口内容。
func Build(w fyne.Window, prefs fyne.Preferences) fyne.CanvasObject {
	initialTheme(prefs)
	cfg := config.Load(prefs)

	var chrome []themedRect
	trackRect := func(name fyne.ThemeColorName) *canvas.Rectangle {
		r := canvas.NewRectangle(currentTheme().Color(name, currentTheme().Variant))
		chrome = append(chrome, themedRect{rect: r, name: name})
		return r
	}

	// ---- 设置区（字段与原实现一致） ----
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetText(cfg.Token)
	ownerEntry := widget.NewEntry()
	ownerEntry.SetText(cfg.Owner)
	repoEntry := widget.NewEntry()
	repoEntry.SetText(cfg.Repo)
	branchEntry := widget.NewEntry()
	branchEntry.SetText(cfg.Branch)
	dirEntry := widget.NewEntry()
	dirEntry.SetText(cfg.Dir)
	intervalEntry := widget.NewEntry()
	intervalEntry.SetText(strconv.Itoa(cfg.IntervalSec))
	retriesEntry := widget.NewEntry()
	retriesEntry.SetText(strconv.Itoa(cfg.Retries))
	autoCreateCheck := widget.NewCheck("仓库不存在时自动创建", nil)
	autoCreateCheck.SetChecked(cfg.AutoCreate)
	tokenStatus := widget.NewLabel("")

	readConfig := func() config.Config {
		iv, _ := strconv.Atoi(strings.TrimSpace(intervalEntry.Text))
		rt, _ := strconv.Atoi(strings.TrimSpace(retriesEntry.Text))
		c := config.Config{
			Token:       strings.TrimSpace(tokenEntry.Text),
			Owner:       strings.TrimSpace(ownerEntry.Text),
			Repo:        strings.TrimSpace(repoEntry.Text),
			Branch:      strings.TrimSpace(branchEntry.Text),
			Dir:         strings.TrimSpace(dirEntry.Text),
			AutoCreate:  autoCreateCheck.Checked,
			IntervalSec: iv,
			Retries:     rt,
		}
		c.Normalize()
		return c
	}

	validateBtn := widget.NewButton("验证 Token", func() {
		c := readConfig()
		if c.Token == "" {
			dialog.ShowError(fmt.Errorf("请先填写 Token"), w)
			return
		}
		tokenStatus.SetText("验证中…")
		go func() {
			login, err := github.New(c.Token).ValidateToken(context.Background())
			fyne.Do(func() {
				if err != nil {
					tokenStatus.SetText("无效: " + err.Error())
				} else {
					tokenStatus.SetText("✓ 已登录: " + login)
				}
			})
		}()
	})

	settingsForm := widget.NewForm(
		widget.NewFormItem("Token", container.NewBorder(nil, nil, nil, validateBtn, tokenEntry)),
		widget.NewFormItem("", tokenStatus),
		widget.NewFormItem("Owner", ownerEntry),
		widget.NewFormItem("Repo", repoEntry),
		widget.NewFormItem("分支", branchEntry),
		widget.NewFormItem("目标目录", dirEntry),
		widget.NewFormItem("发布间隔(秒)", intervalEntry),
		widget.NewFormItem("失败重试次数", retriesEntry),
		widget.NewFormItem("", autoCreateCheck),
	)
	settingsCard := widget.NewCard("发布设置", "", settingsForm)

	// ---- 队列表 ----
	var arts []article.Article
	var rowState *RowState
	cellPool := map[fyne.CanvasObject]*tableCell{}

	table := widget.NewTable(
		func() (int, int) { return len(arts) + 1, 3 },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("cell")
			b := newBadge()
			root := container.NewStack(lbl, b.object())
			cellPool[root] = &tableCell{label: lbl, badge: b}
			return root
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			cell := cellPool[o]
			if id.Row == 0 {
				cell.badge.object().Hide()
				cell.label.Show()
				cell.label.TextStyle = fyne.TextStyle{Bold: true}
				cell.label.SetText([]string{"文件名", "仓库路径", "状态"}[id.Col])
				return
			}
			a := arts[id.Row-1]
			switch id.Col {
			case 0:
				cell.badge.object().Hide()
				cell.label.Show()
				cell.label.TextStyle = fyne.TextStyle{}
				cell.label.SetText(a.Title)
			case 1:
				cell.badge.object().Hide()
				cell.label.Show()
				cell.label.TextStyle = fyne.TextStyle{}
				cell.label.SetText(a.RepoPath)
			case 2:
				cell.label.Hide()
				state := "待发布"
				if rowState != nil {
					state = rowState.Status(id.Row - 1)
				}
				cell.badge.set(currentTheme(), state)
				cell.badge.object().Show()
			}
		},
	)
	table.SetColumnWidth(0, 220)
	table.SetColumnWidth(1, 280)
	table.SetColumnWidth(2, 220)

	logPanel := NewLogPanel(currentTheme())
	pill := newStatusPill()

	elapsedLabel := widget.NewLabel("00:00")
	totalLabel := widget.NewLabel("0")
	successLabel := widget.NewLabel("0")
	failLabel := widget.NewLabel("0")
	pendingLabel := widget.NewLabel("0")
	var startTime time.Time

	refreshCounts := func() {
		total := len(arts)
		success, fail := 0, 0
		if rowState != nil {
			for i := 0; i < total; i++ {
				switch {
				case rowState.Status(i) == "成功":
					success++
				case strings.HasPrefix(rowState.Status(i), "失败"):
					fail++
				}
			}
		}
		totalLabel.SetText(strconv.Itoa(total))
		successLabel.SetText(strconv.Itoa(success))
		failLabel.SetText(strconv.Itoa(fail))
		pendingLabel.SetText(strconv.Itoa(total - success - fail))
	}

	intervalSummary := widget.NewLabel("0 秒")
	retriesSummary := widget.NewLabel("0 次")
	queueSummary := widget.NewLabel("0 篇")
	refreshRightPanel := func() {
		c := readConfig()
		intervalSummary.SetText(fmt.Sprintf("%d 秒", c.IntervalSec))
		retriesSummary.SetText(fmt.Sprintf("%d 次", c.Retries))
		queueSummary.SetText(fmt.Sprintf("%d 篇", len(arts)))
	}
	intervalEntry.OnChanged = func(string) { refreshRightPanel() }
	retriesEntry.OnChanged = func(string) { refreshRightPanel() }

	progress := widget.NewLabel("0/0")

	reloadQueue := func(paths []string) {
		c := readConfig()
		list, err := article.ScanPaths(paths, c.Dir)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		arts = list
		rowState = nil
		progress.SetText(fmt.Sprintf("0/%d", len(arts)))
		refreshCounts()
		refreshRightPanel()
		table.Refresh()
	}

	addFolderBtn := widget.NewButton("选择文件夹", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			reloadQueue([]string{uri.Path()})
		}, w)
	})
	addFileBtn := widget.NewButton("添加文件", func() {
		dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			defer rc.Close()
			reloadQueue([]string{rc.URI().Path()})
		}, w)
	})
	clearBtn := widget.NewButton("清空", func() {
		arts = nil
		rowState = nil
		progress.SetText("0/0")
		refreshCounts()
		refreshRightPanel()
		table.Refresh()
	})

	// ---- 控制区 ----
	var cancelFn context.CancelFunc
	var isRunning bool
	startBtn := widget.NewButton("开始发布", nil)
	startBtn.Importance = widget.HighImportance
	stopBtn := widget.NewButton("停止", nil)
	stopBtn.Importance = widget.DangerImportance
	stopBtn.Disable()
	exportBtn := widget.NewButton("导出链接列表", nil)

	setRunning := func(running bool) {
		isRunning = running
		if running {
			startBtn.Disable()
			stopBtn.Enable()
		} else {
			startBtn.Enable()
			stopBtn.Disable()
		}
		total, done, failed := 0, 0, 0
		if rowState != nil {
			total = len(arts)
			done = rowState.Done()
			for i := 0; i < total; i++ {
				if strings.HasPrefix(rowState.Status(i), "失败") {
					failed++
				}
			}
		}
		pill.repaint(running, done, total, failed)
	}

	startBtn.OnTapped = func() {
		c := readConfig()
		if err := c.Validate(); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if len(arts) == 0 {
			dialog.ShowError(fmt.Errorf("队列为空，请先添加文件"), w)
			return
		}
		c.Save(prefs)

		rowState = NewRowState(len(arts))
		table.Refresh()
		startTime = time.Now()
		elapsedLabel.SetText("00:00")
		setRunning(true)
		logPanel.AppendInfo(fmt.Sprintf("开始发布 %d 篇到 %s/%s", len(arts), c.Owner, c.Repo))

		ctx, cancel := context.WithCancel(context.Background())
		cancelFn = cancel

		go func() {
			client := github.New(c.Token)
			exists, err := client.RepoExists(ctx, c.Owner, c.Repo)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, w)
					setRunning(false)
				})
				return
			}
			if !exists {
				if !c.AutoCreate {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("仓库 %s/%s 不存在（可勾选自动创建）", c.Owner, c.Repo), w)
						setRunning(false)
					})
					return
				}
				if err := client.CreateRepo(ctx, c.Repo); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("创建仓库失败: %w", err), w)
						setRunning(false)
					})
					return
				}
				fyne.Do(func() { logPanel.AppendInfo("已自动创建仓库 " + c.Repo) })
			}

			p := publisher.New(github.NewAdapter(client))
			runErr := p.Run(ctx, c, arts, func(e publisher.Event) {
				fyne.Do(func() {
					rowState.Apply(e)
					progress.SetText(rowState.Progress())
					refreshCounts()
					table.Refresh()
					logPanel.AppendEvent(e)
					elapsedLabel.SetText(formatElapsed(time.Since(startTime)))
					total := len(arts)
					failed := 0
					for i := 0; i < total; i++ {
						if strings.HasPrefix(rowState.Status(i), "失败") {
							failed++
						}
					}
					pill.repaint(true, rowState.Done(), total, failed)
				})
			})
			fyne.Do(func() {
				elapsedLabel.SetText(formatElapsed(time.Since(startTime)))
				if runErr != nil {
					logPanel.AppendInfo("已停止: " + runErr.Error())
				} else {
					logPanel.AppendInfo("全部完成")
				}
				setRunning(false)
			})
		}()
	}

	stopBtn.OnTapped = func() {
		if cancelFn != nil {
			cancelFn()
			logPanel.AppendInfo("正在停止…")
		}
	}

	exportBtn.OnTapped = func() {
		if rowState == nil {
			dialog.ShowInformation("提示", "还没有发布结果", w)
			return
		}
		urls := rowState.SuccessURLs()
		if len(urls) == 0 {
			dialog.ShowInformation("提示", "暂无成功链接", w)
			return
		}
		dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			_, _ = wc.Write([]byte(strings.Join(urls, "\n")))
		}, w)
	}

	saveBtn := widget.NewButton("保存配置", func() {
		readConfig().Save(prefs)
		logPanel.AppendInfo("配置已保存")
	})

	// ---- 日志面板工具条 ----
	autoScrollCheck := widget.NewCheck("自动滚动", func(on bool) { logPanel.SetAutoScroll(on) })
	autoScrollCheck.SetChecked(true)
	copyLogBtn := widget.NewButton("复制", func() {
		fyne.CurrentApp().Clipboard().SetContent(logPanel.Text())
	})
	exportLogBtn := widget.NewButton("导出", func() {
		dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			_, _ = wc.Write([]byte(logPanel.Text()))
		}, w)
	})
	clearLogBtn := widget.NewButton("清空", func() { logPanel.Clear() })

	logHeaderBg := trackRect(theme.ColorNameMenuBackground)
	logHeader := container.NewStack(logHeaderBg, container.NewPadded(container.NewHBox(
		widget.NewLabelWithStyle("运行日志 RUNTIME LOG", fyne.TextAlignLeading, fyne.TextStyle{}),
		layout.NewSpacer(),
		autoScrollCheck, copyLogBtn, exportLogBtn, clearLogBtn,
	)))
	logBg := canvas.NewRectangle(LogBackground)
	logArea := container.NewStack(logBg, logPanel.CanvasObject())
	logSection := container.NewBorder(logHeader, nil, nil, nil, logArea)

	// ---- 状态栏 ----
	statusBarBg := trackRect(theme.ColorNameMenuBackground)
	statusBar := container.NewStack(statusBarBg, container.NewPadded(container.NewHBox(
		widget.NewLabel("总数"), totalLabel,
		widget.NewLabel("成功"), successLabel,
		widget.NewLabel("失败"), failLabel,
		widget.NewLabel("待发"), pendingLabel,
		layout.NewSpacer(),
		widget.NewLabel("已用时"), elapsedLabel,
	)))

	// ---- 发布页 / 帮助页 ----
	toolbar := container.NewHBox(addFolderBtn, addFileBtn, clearBtn, layout.NewSpacer(), saveBtn)
	publishView := container.NewBorder(container.NewVBox(toolbar, settingsCard), nil, nil, nil, table)
	helpView := BuildHelp()
	contentStack := container.NewStack(publishView, helpView)
	helpView.Hide()

	navPublish := widget.NewButton("发布", nil)
	navHelp := widget.NewButton("帮助", nil)
	selectNav := func(active *widget.Button) {
		for _, b := range []*widget.Button{navPublish, navHelp} {
			if b == active {
				b.Importance = widget.HighImportance
			} else {
				b.Importance = widget.LowImportance
			}
			b.Refresh()
		}
	}
	navPublish.OnTapped = func() {
		helpView.Hide()
		publishView.Show()
		selectNav(navPublish)
	}
	navHelp.OnTapped = func() {
		publishView.Hide()
		helpView.Show()
		selectNav(navHelp)
	}
	selectNav(navPublish)

	navBg := trackRect(theme.ColorNameMenuBackground)
	nav := container.NewStack(navBg, container.NewPadded(container.NewVBox(navPublish, navHelp)))

	// ---- 右侧参数栏 ----
	rightBg := trackRect(theme.ColorNameMenuBackground)
	right := container.NewStack(rightBg, container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("运行参数", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, widget.NewLabel("发布间隔"), intervalSummary),
		container.NewGridWithColumns(2, widget.NewLabel("失败重试"), retriesSummary),
		container.NewGridWithColumns(2, widget.NewLabel("队列篇数"), queueSummary),
		widget.NewSeparator(),
		startBtn, stopBtn,
		widget.NewSeparator(),
		saveBtn, clearBtn, exportBtn,
	)))

	// ---- 标题栏 ----
	iconBg := canvas.NewRectangle(currentTheme().accentColor())
	iconBg.CornerRadius = 6
	iconText := canvas.NewText("G", color.White)
	iconText.TextStyle = fyne.TextStyle{Bold: true}
	iconText.Alignment = fyne.TextAlignCenter
	iconBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(24, 24)), container.NewStack(iconBg, container.NewCenter(iconText)))
	appName := widget.NewLabelWithStyle("GitHub 文章发布器", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	var themeToggleBtn *widget.Button
	repaintChrome := func() {
		for _, c := range chrome {
			c.rect.FillColor = currentTheme().Color(c.name, currentTheme().Variant)
			c.rect.Refresh()
		}
		iconBg.FillColor = currentTheme().accentColor()
		iconBg.Refresh()
		table.Refresh()

		total, done, failed := len(arts), 0, 0
		if rowState != nil {
			done = rowState.Done()
			for i := 0; i < total; i++ {
				if strings.HasPrefix(rowState.Status(i), "失败") {
					failed++
				}
			}
		}
		pill.repaint(isRunning, done, total, failed)

		if currentTheme().Variant == theme.VariantDark {
			themeToggleBtn.SetText("☀ 浅色")
		} else {
			themeToggleBtn.SetText("🌙 深色")
		}
	}

	themeToggleBtn = widget.NewButton("", func() {
		next := theme.VariantLight
		if currentTheme().Variant == theme.VariantLight {
			next = theme.VariantDark
		}
		variantName := "light"
		if next == theme.VariantDark {
			variantName = "dark"
		}
		prefs.SetString(keyThemeVariant, variantName)
		fyne.CurrentApp().Settings().SetTheme(NewTheme(next))
	})
	fyne.CurrentApp().Settings().AddListener(func(fyne.Settings) { repaintChrome() })
	repaintChrome()

	titleBarBg := trackRect(theme.ColorNameMenuBackground)
	titleBar := container.NewStack(titleBarBg, container.NewPadded(container.NewHBox(
		iconBox, appName, layout.NewSpacer(), pill.object(), themeToggleBtn,
	)))

	refreshRightPanel()
	refreshCounts()

	body := container.NewBorder(nil, nil, nav, right, contentStack)
	return container.NewBorder(titleBar, container.NewVBox(logSection, statusBar), nil, nil, body)
}
```

- [ ] **Step 2: 编译**

Run: `go build ./...`
Expected: 无输出，退出码 0（若报错，多半是拼写/import 顺序问题，对照上面代码逐处检查）

- [ ] **Step 3: 跑现有单元测试，确认 `RowState` 相关测试未受影响**

Run: `go test ./... -v`
Expected: `internal/ui` 的 `TestApplyEvent` 以及 Task 1–4 新增的全部测试 PASS；`internal/config`、`internal/github`、`internal/article`、`internal/publisher` 测试保持 PASS（本任务未改动这些包）

- [ ] **Step 4: 手动运行验证**

Run: `go run .`

逐项检查（这是本任务里唯一没有自动化测试覆盖的部分，必须手动确认）：
1. 窗口打开后能看到标题栏（图标+"GitHub 文章发布器"+状态胶囊"待机"+主题切换按钮）、左侧"发布/帮助"两个导航项、发布设置卡片、空队列表格、右侧参数栏、底部日志面板与状态栏。
2. 点击左侧"帮助"，内容区切换为 README 摘要卡片列表；点击"发布"切回设置+队列表。
3. 点击标题栏主题切换按钮，整体配色在 Light/Dark 之间切换（标题栏/左导航/右侧栏/日志面板头/状态栏背景，以及队列表状态徽标颜色都应跟着变）。
4. 「选择文件夹」选一个含 `.md` 文件的目录，队列表出现文件，右侧"队列篇数"更新，状态栏"总数"更新。
5. 填入一个真实（或先用无效值验证报错路径）Token/Owner/Repo 后点击「验证 Token」，状态文字正确显示验证中/成功/失败。
6. 点击「开始发布」：状态胶囊变为"发布中 · N/M"（琥珀/蓝色），日志面板出现带时间戳、彩色标签的行，状态栏"已用时"开始变化，队列表状态列出现彩色徽标（发布中=黄、成功=绿、失败=红）。
7. 发布过程中点击「停止」，能正常取消，胶囊与按钮状态恢复正常。
8. 全部完成后点击「导出链接列表」能弹出保存对话框并写出成功链接。
9. 日志面板工具条：取消勾选"自动滚动"后新日志不再自动滚到底部；点击「复制」后粘贴板能取到日志文本；点击「导出」能保存为 txt；点击「清空」日志区变空。
10. 关闭重开程序，之前手动切换过的主题、上次填写的设置（含 Token）应保持（沿用原有 Preferences 持久化逻辑）。

如发现某一项与预期不符，先定位是布局问题还是逻辑问题，修正后重新走一遍本 Step，再进入下一步。

- [ ] **Step 5: 提交**

```bash
git add internal/ui/mainwindow.go
git commit -m "$(cat <<'EOF'
feat(ui): 重写主窗口为设计稿视觉外壳

标题栏(图标/状态胶囊/主题切换) + 左侧导航(发布/帮助) +
发布页(设置卡片+队列表，状态列用彩色徽标) + 右侧参数栏
(只读摘要+开始停止+快捷按钮) + 终端风格日志面板 + 底部
状态栏。发布/验证/导出等业务逻辑与原实现保持一致。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 7: 调整 `main.go` 默认窗口尺寸

**Files:**
- Modify: `main.go:14`

- [ ] **Step 1: 把默认窗口尺寸从 760×640 改为 1120×760**

`main.go` 第 14 行，把：

```go
	w.Resize(fyne.NewSize(760, 640))
```

改为：

```go
	w.Resize(fyne.NewSize(1120, 760))
```

- [ ] **Step 2: 编译**

Run: `go build ./...`
Expected: 无输出，退出码 0

- [ ] **Step 3: 提交**

```bash
git add main.go
git commit -m "$(cat <<'EOF'
feat: 默认窗口尺寸调整为 1120x760

配合新外壳(标题栏+左右侧栏+日志面板)给出更合适的默认可视区域。

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01W7UN21Y7RoGdi1w8ohFiAZ
EOF
)"
```

---

### Task 8: 整体验证

**Files:** 无新增/修改，仅验证。

- [ ] **Step 1: 完整测试套件**

Run: `go test ./... -v`
Expected: 全部 PASS（`internal/config`、`internal/github`、`internal/article`、`internal/publisher`、`internal/ui` 共计原有 + 本计划新增的用例）

- [ ] **Step 2: `go vet`**

Run: `go vet ./...`
Expected: 无输出

- [ ] **Step 3: 构建产物（对齐 README 的构建方式）**

Run: `go build -o /tmp/ghpublisher-verify .`
Expected: 无输出，退出码 0；`/tmp/ghpublisher-verify` 生成成功后可删除

- [ ] **Step 4: 复核 Task 6 Step 4 的手动验证清单是否仍然全部通过**

如果 Task 7 改了窗口尺寸后有布局挤压/溢出（尤其右侧参数栏和日志面板在较小分辨率下），用 `go run .` 再过一遍，视需要微调间距，微调后重新执行 Step 1–3。

本任务无需提交（除非 Step 4 发现问题并修复，那就按 Task 6 的提交格式再提交一次修复）。
