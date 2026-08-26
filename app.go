package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gitmd/internal/config"
	"gitmd/internal/configdir"
	"gitmd/internal/contentgen"
	"gitmd/internal/contentstore"
	"gitmd/internal/github"
	"gitmd/internal/proxycheck"
	"gitmd/internal/publish"
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

	// 批量发布任务的生命周期：StartPublish 起、StopPublish 取消。
	publishMu     sync.Mutex
	publishing    bool
	publishCancel context.CancelFunc
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

// SaveTemplates 只把标题模板与正文模板 A/B 写回素材目录；词库改由「文件库」直接改文件，
// 所以这里不动它们，避免覆盖用户在文件库里的编辑。出错时返回错误文案。
func (a *App) SaveTemplates(title string, bodies []string) string {
	dir, err := contentstore.DefaultDir()
	if err != nil {
		return err.Error()
	}
	if err := contentstore.SaveTemplates(dir, title, bodies); err != nil {
		return err.Error()
	}
	return ""
}

// OpenContentDir 在系统文件管理器里打开素材目录，方便直接用文本编辑器改 txt。
// 直接调系统的 open/explorer/xdg-open，比 file:// URL 更可靠（路径里有空格/中文也不怕）。
func (a *App) OpenContentDir() string {
	dir, err := a.contentDir() // contentDir 会 ensureLayout：目录不存在时先建出来
	if err != nil {
		return err.Error()
	}
	if err := openInFileManager(dir); err != nil {
		return err.Error()
	}
	a.emitInfo("已打开素材目录：" + dir)
	return ""
}

// openInFileManager 用当前系统的文件管理器打开一个目录。
func openInFileManager(path string) error {
	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	// 只负责唤起，不等它退出（explorer 成功也会返回非 0）。
	return cmd.Start()
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

// ---------- 素材目录（文件库浏览与编辑） ----------
//
// 「文件库」直接浏览、编辑生成用的素材目录（contentstore.DefaultDir，即
// <用户配置>/gitmd/content）。目录不存在时先建出带示例的骨架再读，保证第一次
// 打开就有东西可编辑。

// contentDir 返回素材目录，并顺便建出缺失的骨架文件。
func (a *App) contentDir() (string, error) {
	dir, err := contentstore.DefaultDir()
	if err != nil {
		return "", err
	}
	// Load 会 ensureLayout：目录/文件不存在时先建出来。
	if _, err := contentstore.Load(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// ListConfigTree 读取素材目录下的文件树，供「文件库」标签页展示。
func (a *App) ListConfigTree() ([]configdir.Node, error) {
	dir, err := a.contentDir()
	if err != nil {
		return nil, err
	}
	return configdir.Tree(dir)
}

// ReadConfigFile 读取素材目录下某个文件的内容用于预览；过大的文件会被截断。
func (a *App) ReadConfigFile(relPath string) (configdir.FilePreview, error) {
	dir, err := a.contentDir()
	if err != nil {
		return configdir.FilePreview{}, err
	}
	return configdir.ReadFile(dir, relPath)
}

// WriteConfigFile 把「文件库」里编辑后的内容写回素材目录对应的文件。
func (a *App) WriteConfigFile(relPath, content string) string {
	dir, err := a.contentDir()
	if err != nil {
		return err.Error()
	}
	if err := configdir.WriteFile(dir, relPath, content); err != nil {
		return err.Error()
	}
	return ""
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
	// {日期N}/{时间N} 按北京时间（UTC+8）输出；用 FixedZone 免依赖各平台 tzdata。
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	drafts, err := contentgen.Generate(lib, opts, rnd, now)
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

// ---------- 代理 ----------

// ProxyTestResult 是代理连通性拨测结果，回给前端展示。
type ProxyTestResult struct {
	Ok         bool   `json:"ok"`         // 是否连通（拿到任何 HTTP 响应即 true）
	Message    string `json:"message"`    // 展示文案（成功/失败）
	StatusCode int    `json:"statusCode"` // 目标返回的状态码，0 表示未拿到
	LatencyMs  int64  `json:"latencyMs"`  // 往返耗时（毫秒）
}

// TestProxy 走给定代理拨测 github.com 连通性。proxyURL 支持 socks5:// / http:// / https://
// （可带 user:pass@），空串表示直连。拿到任何 HTTP 响应即算连通。
func (a *App) TestProxy(proxyURL string) ProxyTestResult {
	res, err := proxycheck.Check(a.ctx, proxyURL)
	if err != nil {
		return ProxyTestResult{Ok: false, Message: "不通：" + err.Error()}
	}
	ms := res.Latency.Milliseconds()
	return ProxyTestResult{
		Ok:         true,
		StatusCode: res.StatusCode,
		LatencyMs:  ms,
		Message:    fmt.Sprintf("连通 · HTTP %d · %d ms", res.StatusCode, ms),
	}
}

// AccountCheckResult 是账号验活结果，回给前端。
type AccountCheckResult struct {
	Ok        bool   `json:"ok"`        // 活号且请求成功
	Bad       bool   `json:"bad"`       // 明确坏号（鉴权失败 401/403）
	RepoCount int    `json:"repoCount"` // 活号时的仓库数
	Message   string `json:"message"`   // 展示文案
}

// CheckAccount 用账号 CK（整串会话 Cookie）+ 给定代理，真实调 GitHub ListRepos 验活。
// proxyURL 空串表示直连。成功=活号并回显仓库数；401/403=坏号；其它错误=失败（可能是代理/网络）。
func (a *App) CheckAccount(ck, proxyURL string) AccountCheckResult {
	ck = strings.TrimSpace(ck)
	if ck == "" {
		return AccountCheckResult{Message: "CK 为空"}
	}

	c, err := github.New(ck, github.WithProxy(proxyURL))
	if err != nil {
		return AccountCheckResult{Message: "构造客户端失败：" + err.Error()}
	}

	resp, err := c.ListRepos(a.ctx, 1)
	repoCount := 0
	if resp != nil {
		repoCount = resp.Payload.ReposFinderPageRoute.RepositoryCount
	}
	return accountVerdict(repoCount, c.LastStatusCode(), err)
}

// accountVerdict 把 ListRepos 的结果归类为 活号/坏号/失败（纯函数，便于单测）。
func accountVerdict(repoCount, statusCode int, err error) AccountCheckResult {
	if err == nil {
		return AccountCheckResult{Ok: true, RepoCount: repoCount, Message: fmt.Sprintf("活号 · 仓库 %d 个", repoCount)}
	}
	if statusCode == 401 || statusCode == 403 {
		return AccountCheckResult{Bad: true, Message: fmt.Sprintf("坏号 · HTTP %d", statusCode)}
	}
	return AccountCheckResult{Message: "失败：" + err.Error()}
}

// ---------- 批量发布 ----------

// PublishAccount 是前端传入的一个待发布账号。
type PublishAccount struct {
	ID int    `json:"id"`
	CK string `json:"ck"`
}

// PublishConfig 是前端传入的发布配置（本轮生效的核心参数）。
type PublishConfig struct {
	Threads       int    `json:"threads"`       // 线程数量
	Interval      int    `json:"interval"`      // 发布间隔（秒）
	PerAccount    int    `json:"perAccount"`    // 每号发布次数
	FailSwitch    int    `json:"failSwitch"`    // 失败换号
	AccountCycles int    `json:"accountCycles"` // 账号循环轮数
	RoundInterval int    `json:"roundInterval"` // 每轮间隔（秒）
	ProxyURL      string `json:"proxyUrl"`      // 代理（启用时非空）
}

// PublishAccountUpdate 是发布过程中回给前端的账号状态更新（经 publish:account 事件）。
type PublishAccountUpdate struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	Success int    `json:"success"`
	Fail    int    `json:"fail"`
}

// PublishLink 是一条发布成功的链接（经 publish:link 事件回前端）。
type PublishLink struct {
	ID   int    `json:"id"`   // 账号 id
	Repo string `json:"repo"` // 仓库名
	File string `json:"file"` // 文件名
	URL  string `json:"url"`  // GitHub 文件链接
}

// 发布事件名。
const (
	eventPublishAccount = "publish:account" // 单账号状态更新
	eventPublishLink    = "publish:link"    // 发布成功一篇的链接
	eventPublishDone    = "publish:done"    // 整个任务结束
)

// StartPublish 启动一次批量发布：加载素材库 → 起 worker-pool 引擎（在后台协程跑），
// 进度经事件回前端。立即返回；出错返回错误文案，成功返回空串。
func (a *App) StartPublish(accounts []PublishAccount, cfg PublishConfig, genOpts contentgen.Options) string {
	a.publishMu.Lock()
	if a.publishing {
		a.publishMu.Unlock()
		return "已有发布任务在运行"
	}

	dir, err := contentstore.DefaultDir()
	if err != nil {
		a.publishMu.Unlock()
		return err.Error()
	}
	lib, err := contentstore.Load(dir)
	if err != nil {
		a.publishMu.Unlock()
		return err.Error()
	}

	accs := make([]publish.Account, 0, len(accounts))
	for _, x := range accounts {
		ck := strings.TrimSpace(x.CK)
		if ck == "" {
			continue
		}
		accs = append(accs, publish.Account{ID: x.ID, CK: ck})
	}
	if len(accs) == 0 {
		a.publishMu.Unlock()
		return "没有可发布的账号"
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.publishCancel = cancel
	a.publishing = true
	a.publishMu.Unlock()

	runner := publish.New(publish.Config{
		Threads:          cfg.Threads,
		IntervalSec:      cfg.Interval,
		PerAccount:       cfg.PerAccount,
		FailSwitch:       cfg.FailSwitch,
		Cycles:           cfg.AccountCycles,
		RoundIntervalSec: cfg.RoundInterval,
		ProxyURL:         cfg.ProxyURL,
	}, accs, lib, genOpts, wailsReporter{a: a}, time.Now().UnixNano())

	a.emitInfo(fmt.Sprintf("开始发布：账号 %d 个 · 线程 %d · 每号 %d 篇 · %d 轮", len(accs), cfg.Threads, cfg.PerAccount, max(cfg.AccountCycles, 1)))

	go func() {
		defer func() {
			a.publishMu.Lock()
			a.publishing = false
			a.publishCancel = nil
			a.publishMu.Unlock()
			runtime.EventsEmit(a.ctx, eventPublishDone, nil)
		}()
		runner.Run(ctx)
	}()
	return ""
}

// StopPublish 取消正在运行的发布任务；无任务时为 no-op。
func (a *App) StopPublish() {
	a.publishMu.Lock()
	if a.publishCancel != nil {
		a.publishCancel()
	}
	a.publishMu.Unlock()
}

// LinksFileName 是「查看链接」输出到素材目录根下的文件名，与「外链库.txt」区分开。
const LinksFileName = "查看链接.txt"

// WriteLinksFile 把发布链接写入素材目录下的 查看链接.txt（不存在则创建）。
// 供「查看链接」按钮先落盘、再在文件库里打开。返回错误文案或空串。
func (a *App) WriteLinksFile(content string) string {
	dir, err := a.contentDir() // contentDir 会 ensureLayout：目录不存在时先建
	if err != nil {
		return err.Error()
	}
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if err := os.WriteFile(filepath.Join(dir, LinksFileName), []byte(normalized), 0o644); err != nil {
		return err.Error()
	}
	return ""
}

// EnsureLinksFile 确保 查看链接.txt 存在（不存在则建空文件，已存在则不动，保留上次的链接）。
// 供「查看链接」在本次没有新链接时也能跳转打开该文件。返回错误文案或空串。
func (a *App) EnsureLinksFile() string {
	dir, err := a.contentDir()
	if err != nil {
		return err.Error()
	}
	p := filepath.Join(dir, LinksFileName)
	if _, err := os.Stat(p); err == nil {
		return "" // 已存在，保留内容
	}
	if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
		return err.Error()
	}
	return ""
}

// wailsReporter 把 publish 引擎的进度转成 Wails 事件回前端。EventsEmit 并发安全。
type wailsReporter struct{ a *App }

func (r wailsReporter) Log(kind, tag, msg string) {
	runtime.EventsEmit(r.a.ctx, eventLog, LogLine{
		Time: time.Now().Format("15:04:05"), Tag: tag, Kind: kind, Msg: msg,
	})
}

func (r wailsReporter) Account(id int, status string, success, fail int) {
	runtime.EventsEmit(r.a.ctx, eventPublishAccount, PublishAccountUpdate{
		ID: id, Status: status, Success: success, Fail: fail,
	})
}

func (r wailsReporter) Published(id int, repo, file, url string) {
	runtime.EventsEmit(r.a.ctx, eventPublishLink, PublishLink{
		ID: id, Repo: repo, File: file, URL: url,
	})
}

// ---------- 通用工具 ----------

// CopyToClipboard 把文本写入系统剪贴板。
func (a *App) CopyToClipboard(text string) string {
	if err := runtime.ClipboardSetText(a.ctx, text); err != nil {
		return err.Error()
	}
	return ""
}

// ClipboardGetText 读取系统剪贴板里的纯文本，供发布页「双击粘贴剪贴板」批量导入账号。
// 读不到内容时返回空字符串。
func (a *App) ClipboardGetText() string {
	text, err := runtime.ClipboardGetText(a.ctx)
	if err != nil {
		return ""
	}
	return text
}

// SaveTextFile 弹出保存框，把一段纯文本写入用户选择的文件，供发布页「导出结果」使用。
// 用户取消或成功都返回空字符串，出错时返回错误文案。
func (a *App) SaveTextFile(defaultName, text string) string {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "导出结果", DefaultFilename: defaultName,
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
	a.emitInfo("已导出到 " + path)
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
