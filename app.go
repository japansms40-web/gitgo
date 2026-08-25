package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
	"githubbaidu/internal/github"
	"githubbaidu/internal/publisher"
	"githubbaidu/internal/runlog"
)

const (
	eventLog    = "publish:log"    // 一条运行日志（LogLine）
	eventStatus = "publish:status" // 队列里某一行的状态变化（StatusUpdate）
	eventDone   = "publish:done"   // 本轮发布结束，携带错误文案（空字符串=正常结束）
)

// QueueItem 是发给前端展示用的精简文章信息，不含正文内容。
type QueueItem struct {
	Title    string `json:"title"`
	RepoPath string `json:"repoPath"`
}

// LogLine 是发给前端渲染的一条运行日志。
type LogLine struct {
	Time      string `json:"time"`
	Tag       string `json:"tag"`
	Kind      string `json:"kind"`
	Msg       string `json:"msg"`
	Highlight bool   `json:"highlight"`
}

// StatusUpdate 描述队列里第 Index 篇文章的最新状态。
type StatusUpdate struct {
	Index int    `json:"index"`
	Kind  string `json:"kind"`
	URL   string `json:"url"`
	Err   string `json:"err"`
}

// App 是 Wails 绑定给前端调用的方法集合，负责把 internal/* 里的业务逻辑
// 接到前端；本身只做编排和事件转发，不实现业务规则。
type App struct {
	ctx context.Context

	arts []article.Article

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// LoadConfig 读取磁盘上保存的配置；不存在则返回默认值。
func (a *App) LoadConfig() config.Config {
	path, err := config.DefaultPath()
	if err != nil {
		c := config.Config{Branch: "main", Dir: "posts", IntervalSec: 1, Retries: 2}
		return c
	}
	return config.Load(path)
}

// SaveConfig 把配置写入磁盘；出错时返回错误文案，成功返回空字符串。
func (a *App) SaveConfig(cfg config.Config) string {
	cfg.Normalize()
	path, err := config.DefaultPath()
	if err != nil {
		return err.Error()
	}
	if err := cfg.Save(path); err != nil {
		return err.Error()
	}
	return ""
}

// ValidateToken 调用 GitHub API 验证 Token，返回登录名。
func (a *App) ValidateToken(token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("请先填写 Token")
	}
	return github.New(token).ValidateToken(a.ctx)
}

// SelectFolder 弹出系统文件夹选择框，取消时返回空字符串。
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择文件夹"})
}

// SelectFiles 弹出系统多选文件框，取消时返回空切片。
func (a *App) SelectFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "添加文件",
		Filters: []runtime.FileFilter{{DisplayName: "Markdown / 文本", Pattern: "*.md;*.txt"}},
	})
}

// ScanQueue 扫描给定路径下的 .md/.txt 文件，重建发布队列并返回展示用列表。
func (a *App) ScanQueue(paths []string, dir string) ([]QueueItem, error) {
	list, err := article.ScanPaths(paths, dir)
	if err != nil {
		return nil, err
	}
	a.arts = list
	items := make([]QueueItem, len(list))
	for i, art := range list {
		items[i] = QueueItem{Title: art.Title, RepoPath: art.RepoPath}
	}
	return items, nil
}

// ClearQueue 清空当前发布队列。
func (a *App) ClearQueue() {
	a.arts = nil
}

// StartPublish 校验配置与队列后异步开始发布，通过事件把日志与状态实时推给前端。
// 返回非空字符串表示未能开始（校验失败/已有任务在运行），前端应据此提示用户。
func (a *App) StartPublish(cfg config.Config) string {
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return "已有发布任务在运行"
	}
	if err := cfg.Validate(); err != nil {
		a.mu.Unlock()
		return err.Error()
	}
	if len(a.arts) == 0 {
		a.mu.Unlock()
		return "队列为空，请先添加文件"
	}
	cfg.Normalize()
	arts := a.arts
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.mu.Unlock()

	go a.runPublish(ctx, cfg, arts)
	return ""
}

func (a *App) runPublish(ctx context.Context, cfg config.Config, arts []article.Article) {
	defer func() {
		a.mu.Lock()
		a.cancel = nil
		a.mu.Unlock()
	}()

	a.emitInfo(fmt.Sprintf("开始发布 %d 篇到 %s/%s", len(arts), cfg.Owner, cfg.Repo))

	client := github.New(cfg.Token)
	exists, err := client.RepoExists(ctx, cfg.Owner, cfg.Repo)
	if err != nil {
		a.emitInfo("检查仓库失败: " + err.Error())
		runtime.EventsEmit(a.ctx, eventDone, err.Error())
		return
	}
	if !exists {
		if !cfg.AutoCreate {
			msg := fmt.Sprintf("仓库 %s/%s 不存在（可勾选自动创建）", cfg.Owner, cfg.Repo)
			a.emitInfo(msg)
			runtime.EventsEmit(a.ctx, eventDone, msg)
			return
		}
		if err := client.CreateRepo(ctx, cfg.Repo); err != nil {
			msg := "创建仓库失败: " + err.Error()
			a.emitInfo(msg)
			runtime.EventsEmit(a.ctx, eventDone, msg)
			return
		}
		a.emitInfo("已自动创建仓库 " + cfg.Repo)
	}

	p := publisher.New(github.NewAdapter(client))
	runErr := p.Run(ctx, cfg, arts, func(e publisher.Event) {
		tag, kind, hi := runlog.TagFor(e.Kind)
		runtime.EventsEmit(a.ctx, eventLog, LogLine{
			Time: time.Now().Format("15:04:05"), Tag: tag, Kind: string(kind),
			Msg: runlog.LineFor(e), Highlight: hi,
		})
		errText := ""
		if e.Err != nil {
			errText = e.Err.Error()
		}
		runtime.EventsEmit(a.ctx, eventStatus, StatusUpdate{Index: e.Index, Kind: string(kind), URL: e.URL, Err: errText})
	})

	if runErr != nil {
		a.emitInfo("已停止: " + runErr.Error())
	} else {
		a.emitInfo("全部完成")
	}
	runtime.EventsEmit(a.ctx, eventDone, "")
}

// StopPublish 取消正在进行的发布任务；没有任务在跑时什么也不做。
func (a *App) StopPublish() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
		a.emitInfo("正在停止…")
	}
}

func (a *App) emitInfo(msg string) {
	runtime.EventsEmit(a.ctx, eventLog, LogLine{
		Time: time.Now().Format("15:04:05"), Tag: "[信息]", Kind: "info", Msg: msg,
	})
}

// ExportLinks 弹出保存框，把成功链接写入用户选择的文件；取消保存返回空字符串。
func (a *App) ExportLinks(urls []string) string {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "导出链接列表", DefaultFilename: "links.txt",
	})
	if err != nil {
		return err.Error()
	}
	if path == "" {
		return ""
	}
	if err := os.WriteFile(path, []byte(strings.Join(urls, "\n")), 0o644); err != nil {
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

// CopyToClipboard 把文本写入系统剪贴板。
func (a *App) CopyToClipboard(text string) string {
	if err := runtime.ClipboardSetText(a.ctx, text); err != nil {
		return err.Error()
	}
	return ""
}
