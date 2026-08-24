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
