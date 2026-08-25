package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gitmd/internal/config"
	"gitmd/internal/contentgen"
	"gitmd/internal/contentstore"
)

// eventLog 是推给前端日志面板的一行运行日志。
const eventLog = "gen:log"

// LogLine 是发给前端渲染的一条日志。
type LogLine struct {
	Time      string `json:"time"`
	Tag       string `json:"tag"`
	Kind      string `json:"kind"`
	Msg       string `json:"msg"`
	Highlight bool   `json:"highlight"`
}

// App 是 Wails 绑定给前端调用的方法集合，负责把 internal/* 里的业务逻辑
// 接到前端；本身只做编排，不实现生成规则。
type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// ---------- 生成参数 ----------

// LoadConfig 读取磁盘上保存的生成参数；不存在则返回默认值。
func (a *App) LoadConfig() contentgen.Options {
	return config.Load()
}

// SaveConfig 把生成参数写入磁盘；出错时返回错误文案，成功返回空字符串。
func (a *App) SaveConfig(opts contentgen.Options) string {
	if err := config.Save(opts); err != nil {
		return err.Error()
	}
	return ""
}

// ---------- 素材（模板与词库 txt） ----------

// LoadContent 读取素材目录里的模板与词库；目录不存在时会先建出空骨架。
func (a *App) LoadContent() (contentgen.Library, error) {
	dir, err := contentstore.DefaultDir()
	if err != nil {
		return contentgen.Library{}, err
	}
	return contentstore.Load(dir)
}

// SaveContent 把模板与词库写回 txt 文件；出错时返回错误文案。
func (a *App) SaveContent(lib contentgen.Library) string {
	dir, err := contentstore.DefaultDir()
	if err != nil {
		return err.Error()
	}
	if err := contentstore.Save(dir, lib); err != nil {
		return err.Error()
	}
	return ""
}

// OpenContentDir 在系统文件管理器里打开素材目录，方便直接用文本编辑器改 txt。
func (a *App) OpenContentDir() string {
	dir, err := contentstore.DefaultDir()
	if err != nil {
		return err.Error()
	}
	// 目录可能还没建出来，先确保存在再打开。
	if _, err := contentstore.Load(dir); err != nil {
		return err.Error()
	}
	runtime.BrowserOpenURL(a.ctx, "file://"+dir)
	return ""
}

// ImportTextFile 让用户选一个 txt 文件并返回其内容，用于往模板/词库输入框里灌数据。
// 用户取消选择时返回空字符串。
func (a *App) ImportTextFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "导入文本",
		Filters: []runtime.FileFilter{{DisplayName: "文本文件", Pattern: "*.txt;*.md"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ---------- 生成与导出 ----------

// Generate 从磁盘上的素材生成草稿。前端会先调 SaveContent 再调它，
// 这样预览到的内容和存在 txt 里的一定是同一份。
func (a *App) Generate(opts contentgen.Options) ([]contentgen.Draft, error) {
	dir, err := contentstore.DefaultDir()
	if err != nil {
		return nil, err
	}
	lib, err := contentstore.Load(dir)
	if err != nil {
		return nil, err
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	drafts, err := contentgen.Generate(lib, opts, rnd, time.Now())
	if err != nil {
		return nil, err
	}
	a.emitInfo(fmt.Sprintf("已生成 %d 篇草稿", len(drafts)))
	return drafts, nil
}

// ExportDrafts 弹出文件夹选择框，把草稿逐篇写成 .md；
// 出错时返回错误文案，用户取消选择或成功都返回空字符串。
func (a *App) ExportDrafts(drafts []contentgen.Draft) string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择导出目录"})
	if err != nil {
		return err.Error()
	}
	if dir == "" {
		return ""
	}
	n, err := contentstore.ExportDrafts(dir, drafts)
	if err != nil {
		return err.Error()
	}
	a.emitInfo(fmt.Sprintf("已导出 %d 篇到 %s", n, dir))
	return ""
}

// ---------- 通用工具 ----------

// CopyToClipboard 把文本写入系统剪贴板。
func (a *App) CopyToClipboard(text string) string {
	if err := runtime.ClipboardSetText(a.ctx, text); err != nil {
		return err.Error()
	}
	return ""
}

// ExportLog 弹出保存框，把前端已渲染的日志纯文本写入用户选择的文件。
func (a *App) ExportLog(text string) string {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "导出日志", DefaultFilename: "run.log",
	})
	if err != nil {
		return err.Error()
	}
	if path == "" {
		return ""
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) emitInfo(msg string) {
	runtime.EventsEmit(a.ctx, eventLog, LogLine{
		Time: time.Now().Format("15:04:05"), Tag: "[信息]", Kind: "info", Msg: msg,
	})
}
