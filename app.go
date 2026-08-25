package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"githubbaidu/internal/account"
	"githubbaidu/internal/accountpublish"
	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
	"githubbaidu/internal/runlog"
)

const (
	eventLog     = "publish:log"     // 一条运行日志（LogLine）
	eventAccount = "publish:account" // 账号队列里某一行的状态变化（AccountUpdate）
	eventRound   = "publish:round"   // 当前轮次进度变化（RoundUpdate）
	eventDone    = "publish:done"    // 本次发布任务结束，携带错误文案（空字符串=正常结束）
)

// QueueItem 是发给前端展示用的精简内容信息，不含正文。
type QueueItem struct {
	Title string `json:"title"`
}

// LogLine 是发给前端渲染的一条运行日志。
type LogLine struct {
	Time      string `json:"time"`
	Tag       string `json:"tag"`
	Kind      string `json:"kind"`
	Msg       string `json:"msg"`
	Highlight bool   `json:"highlight"`
}

// AccountUpdate 描述账号队列里第 Index 个账号的最新状态。
type AccountUpdate struct {
	Index   int    `json:"index"`
	Status  string `json:"status"`
	Success int    `json:"success"`
	Fail    int    `json:"fail"`
	Total   int    `json:"total"`
}

// RoundUpdate 描述当前这一轮账号池的处理进度。
type RoundUpdate struct {
	Round int `json:"round"`
	Done  int `json:"done"`
	Total int `json:"total"`
}

// PublishResult 是一条成功发布的结果记录，供"查看链接"展示。
type PublishResult struct {
	Time  string `json:"time"`
	CK    string `json:"ck"`
	Title string `json:"title"`
	Value string `json:"value"`
}

// App 是 Wails 绑定给前端调用的方法集合，负责把 internal/* 里的业务逻辑
// 接到前端；本身只做编排和事件转发，不实现业务规则。
type App struct {
	ctx context.Context

	arts     []article.Article
	accounts []account.Account
	results  []PublishResult

	mu        sync.Mutex
	cancel    context.CancelFunc
	pauseGate *accountpublish.PauseGate
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// ---------- 任务参数 ----------

// LoadConfig 读取磁盘上保存的任务参数；不存在则返回默认值。
func (a *App) LoadConfig() config.Config {
	return config.Load()
}

// SaveConfig 把任务参数写入磁盘；出错时返回错误文案，成功返回空字符串。
func (a *App) SaveConfig(cfg config.Config) string {
	cfg.Normalize()
	if err := config.Save(cfg); err != nil {
		return err.Error()
	}
	return ""
}

// ---------- 内容队列（本地 Markdown/文本） ----------

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

// ScanQueue 扫描给定路径下的 .md/.txt 文件，重建待发布内容并返回展示用列表。
func (a *App) ScanQueue(paths []string) ([]QueueItem, error) {
	list, err := article.ScanPaths(paths)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.arts = list
	a.mu.Unlock()

	items := make([]QueueItem, len(list))
	for i, art := range list {
		items[i] = QueueItem{Title: art.Title}
	}
	return items, nil
}

// ClearQueue 清空当前待发布内容。
func (a *App) ClearQueue() {
	a.mu.Lock()
	a.arts = nil
	a.mu.Unlock()
}

// ---------- 账号队列（CK/UA/IP） ----------

// SelectAccountFiles 弹出系统多选文件框（仅 .txt），取消时返回空切片。
func (a *App) SelectAccountFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "导入账号",
		Filters: []runtime.FileFilter{{DisplayName: "文本文件", Pattern: "*.txt"}},
	})
}

// LoadAccounts 读取磁盘上保存的账号队列；不存在则返回空列表。
func (a *App) LoadAccounts() []account.Account {
	path, err := account.DefaultPath()
	if err != nil {
		return []account.Account{}
	}
	list := account.Load(path)
	a.mu.Lock()
	a.accounts = list
	a.mu.Unlock()
	return list
}

// saveAccountsLocked 把当前账号队列写入磁盘；调用方需持有 a.mu。
func (a *App) saveAccountsLocked() {
	path, err := account.DefaultPath()
	if err != nil {
		return
	}
	_ = account.Save(path, a.accounts)
}

// ImportAccountsText 解析粘贴/拖入的文本（"----" 分隔）并追加到账号队列。
func (a *App) ImportAccountsText(text string) ([]account.Account, error) {
	parsed := account.ParseImportText(text)
	if len(parsed) == 0 {
		return nil, errors.New("没有解析到有效账号")
	}
	a.mu.Lock()
	a.accounts = append(a.accounts, parsed...)
	a.saveAccountsLocked()
	list := a.accounts
	a.mu.Unlock()
	return list, nil
}

// ImportAccountsFile 读取拖入/选择的文本文件内容，按 "----" 分隔批量导入账号。
func (a *App) ImportAccountsFile(paths []string) ([]account.Account, error) {
	var text strings.Builder
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		text.Write(b)
		text.WriteString("\n----\n")
	}
	return a.ImportAccountsText(text.String())
}

// PasteAccountsFromClipboard 读取系统剪贴板文本并按 "----" 分隔批量导入账号。
func (a *App) PasteAccountsFromClipboard() ([]account.Account, error) {
	text, err := runtime.ClipboardGetText(a.ctx)
	if err != nil {
		return nil, err
	}
	return a.ImportAccountsText(text)
}

// RemoveAccount 把指定下标的账号移出队列。
func (a *App) RemoveAccount(index int) ([]account.Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.accounts) {
		return nil, fmt.Errorf("账号下标越界")
	}
	a.accounts = append(a.accounts[:index], a.accounts[index+1:]...)
	a.saveAccountsLocked()
	return a.accounts, nil
}

// MarkBadAccount 把指定下标的账号标记为坏号；批量发布时会跳过坏号。
func (a *App) MarkBadAccount(index int) ([]account.Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.accounts) {
		return nil, fmt.Errorf("账号下标越界")
	}
	a.accounts[index].Bad = true
	a.saveAccountsLocked()
	return a.accounts, nil
}

// ClearAccounts 清空整个账号队列。
func (a *App) ClearAccounts() []account.Account {
	a.mu.Lock()
	a.accounts = []account.Account{}
	a.saveAccountsLocked()
	list := a.accounts
	a.mu.Unlock()
	return list
}

// ExportAccountsResult 弹出保存框，把账号队列及统计导出为 CSV；取消保存返回空字符串。
func (a *App) ExportAccountsResult() string {
	a.mu.Lock()
	list := append([]account.Account(nil), a.accounts...)
	a.mu.Unlock()

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "导出结果", DefaultFilename: "accounts-result.csv",
	})
	if err != nil {
		return err.Error()
	}
	if path == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("CK,UA,IP,状态,成功,失败,总数\n")
	for _, acc := range list {
		fmt.Fprintf(&b, "%s,%s,%s,%s,%d,%d,%d\n",
			csvField(acc.CK), csvField(acc.UA), csvField(acc.IP), acc.Status, acc.Success, acc.Fail, acc.Total)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err.Error()
	}
	return ""
}

func csvField(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// TestAccount 对单个账号做一次发布尝试（目前走占位实现），用于验证账号是否可用；
// 结果计入该账号的累计统计。
func (a *App) TestAccount(index int) (account.Account, error) {
	a.mu.Lock()
	if index < 0 || index >= len(a.accounts) {
		a.mu.Unlock()
		return account.Account{}, fmt.Errorf("账号下标越界")
	}
	acc := a.accounts[index]
	var art article.Article
	if len(a.arts) > 0 {
		art = a.arts[0]
	}
	a.accounts[index].Status = account.StatusRunning
	a.mu.Unlock()

	_, pubErr := accountpublish.TODORequester{}.Publish(a.ctx, acc, art)

	a.mu.Lock()
	defer a.mu.Unlock()
	if index >= len(a.accounts) { // 测试期间账号被移出列表
		return account.Account{}, fmt.Errorf("账号已被移出列表")
	}
	// pubErr 只是这次测试的结果（成功/失败），不是调用本身出错，
	// 所以这里正常返回 nil error，让前端始终拿到更新后的账号状态。
	a.accounts[index].Total++
	if pubErr != nil {
		a.accounts[index].Fail++
		a.accounts[index].Status = account.StatusFailed
	} else {
		a.accounts[index].Success++
		a.accounts[index].Status = account.StatusSuccess
	}
	a.saveAccountsLocked()
	return a.accounts[index], nil
}

// ---------- 批量发布 ----------

// StartPublish 校验内容与账号队列后异步开始发布，通过事件把日志、账号状态与轮次进度
// 实时推给前端。返回非空字符串表示未能开始（校验失败/已有任务在运行），前端应据此提示用户。
func (a *App) StartPublish(cfg config.Config) string {
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return "已有发布任务在运行"
	}
	if len(a.arts) == 0 {
		a.mu.Unlock()
		return "请先选择要发布的内容"
	}
	var targets []accountpublish.IndexedAccount
	for i, acc := range a.accounts {
		if acc.Bad {
			continue
		}
		targets = append(targets, accountpublish.IndexedAccount{Index: i, Account: acc})
	}
	if len(targets) == 0 {
		a.mu.Unlock()
		return "账号队列为空（或都已标记为坏号），请先导入账号"
	}
	cfg.Normalize()
	arts := a.arts
	ctx, cancel := context.WithCancel(context.Background())
	gate := accountpublish.NewPauseGate()
	a.cancel = cancel
	a.pauseGate = gate
	a.results = nil
	a.mu.Unlock()

	go a.runPublish(ctx, cfg, gate, targets, arts)
	return ""
}

func (a *App) runPublish(ctx context.Context, cfg config.Config, gate *accountpublish.PauseGate, targets []accountpublish.IndexedAccount, arts []article.Article) {
	defer func() {
		a.mu.Lock()
		a.cancel = nil
		a.pauseGate = nil
		a.mu.Unlock()
	}()

	a.emitInfo(fmt.Sprintf("开始发布，账号 %d 个，内容 %d 篇，线程数 %d，每号 %d 次，%d 轮",
		len(targets), len(arts), cfg.Threads, cfg.PerAccountCount, cfg.CycleRounds))

	runConfig := accountpublish.RunConfig{
		Threads:          cfg.Threads,
		IntervalSec:      cfg.IntervalSec,
		PerAccountCount:  cfg.PerAccountCount,
		FailSwitchCount:  cfg.FailSwitchCount,
		CycleRounds:      cfg.CycleRounds,
		RoundIntervalSec: cfg.RoundIntervalSec,
		CreateRepo:       cfg.CreateRepo,
	}
	runner := accountpublish.New(accountpublish.TODORequester{}, accountpublish.TODORepoCreator{})
	runErr := runner.Run(ctx, runConfig, gate, targets, arts, func(e accountpublish.Event) {
		a.handlePublishEvent(e)
	})

	if runErr != nil {
		a.emitInfo("已停止: " + runErr.Error())
	} else {
		a.emitInfo("全部完成")
	}
	runtime.EventsEmit(a.ctx, eventDone, "")
}

// handlePublishEvent 分发一条发布事件：账号相关的落到状态/累计统计上，
// 轮次相关的推送轮次进度，其余都写一行运行日志。
func (a *App) handlePublishEvent(e accountpublish.Event) {
	if e.Kind == accountpublish.EventRoundProgress {
		runtime.EventsEmit(a.ctx, eventRound, RoundUpdate{Round: e.Round, Done: e.RoundDone, Total: e.RoundTotal})
		return
	}

	tag, kind, hi := runlog.TagFor(e.Kind)
	runtime.EventsEmit(a.ctx, eventLog, LogLine{
		Time: time.Now().Format("15:04:05"), Tag: tag, Kind: string(kind),
		Msg: runlog.LineFor(e), Highlight: hi,
	})

	if e.Kind == accountpublish.EventRoundStart {
		runtime.EventsEmit(a.ctx, eventRound, RoundUpdate{Round: e.Round, Done: 0, Total: e.RoundTotal})
		return
	}
	if e.Kind == accountpublish.EventRoundDone {
		return
	}
	a.applyAccountEvent(e)
}

// applyAccountEvent 把一次账号发布事件落到对应账号的状态/累计统计上，并推送给前端。
func (a *App) applyAccountEvent(e accountpublish.Event) {
	a.mu.Lock()
	if e.AccountIndex < 0 || e.AccountIndex >= len(a.accounts) {
		a.mu.Unlock()
		return
	}
	acc := &a.accounts[e.AccountIndex]
	switch e.Kind {
	case accountpublish.EventAttemptStart:
		acc.Status = account.StatusRunning
	case accountpublish.EventAttemptSuccess:
		acc.Status = account.StatusSuccess
		acc.Success++
		acc.Total++
		a.results = append(a.results, PublishResult{
			Time: time.Now().Format("15:04:05"), CK: e.CK, Title: e.ItemLabel, Value: e.Result,
		})
		a.saveAccountsLocked()
	case accountpublish.EventAttemptFailure:
		acc.Status = account.StatusFailed
		acc.Fail++
		acc.Total++
		a.saveAccountsLocked()
	case accountpublish.EventAccountSwitch:
		// 状态在此之前的 Attempt 事件（或建仓库失败）里已经体现，这里不再改状态。
	}
	update := AccountUpdate{Index: e.AccountIndex, Status: acc.Status, Success: acc.Success, Fail: acc.Fail, Total: acc.Total}
	a.mu.Unlock()

	runtime.EventsEmit(a.ctx, eventAccount, update)
}

// PausePublish 暂停正在进行的发布任务；worker 会在当前尝试完成后挂起。
func (a *App) PausePublish() {
	a.mu.Lock()
	gate := a.pauseGate
	a.mu.Unlock()
	if gate != nil {
		gate.Pause()
		a.emitInfo("已暂停")
	}
}

// ResumePublish 恢复被暂停的发布任务。
func (a *App) ResumePublish() {
	a.mu.Lock()
	gate := a.pauseGate
	a.mu.Unlock()
	if gate != nil {
		gate.Resume()
		a.emitInfo("已恢复")
	}
}

// StopPublish 取消正在进行的发布任务；没有任务在跑时什么也不做。
func (a *App) StopPublish() {
	a.mu.Lock()
	cancel := a.cancel
	gate := a.pauseGate
	a.mu.Unlock()
	if cancel != nil {
		cancel()
		a.emitInfo("正在停止…")
	}
	if gate != nil {
		gate.Resume() // 暂停中直接停止的话要唤醒 worker，让它能立即看到 ctx 已取消
	}
}

// GetPublishResults 返回本次发布任务里已经成功的结果列表，供"查看链接"展示。
// 返回值必须是非 nil 切片：append(nil) 在没有元素可加时仍然是 nil，
// 序列化成 JSON 会变成 null，参见 internal/account.Load 上的注释。
func (a *App) GetPublishResults() []PublishResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]PublishResult, len(a.results))
	copy(out, a.results)
	return out
}

func (a *App) emitInfo(msg string) {
	runtime.EventsEmit(a.ctx, eventLog, LogLine{
		Time: time.Now().Format("15:04:05"), Tag: "[信息]", Kind: "info", Msg: msg,
	})
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
