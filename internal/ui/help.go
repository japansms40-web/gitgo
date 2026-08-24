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
