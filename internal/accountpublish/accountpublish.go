// Package accountpublish 编排"用账号队列并发发布内容"的任务：
// 多线程并发处理账号池，单个账号可循环发布多次，连续失败达到阈值换号，
// 整个账号池可循环多轮，支持暂停/恢复与取消，通过回调把进度事件传给上层。
package accountpublish

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"githubbaidu/internal/account"
	"githubbaidu/internal/article"
)

// Requester 执行一次"用某账号发布某篇内容"的请求，返回结果描述或错误。
// 真正的请求协议由外部实现注入；TODORequester 是项目还没接入真实系统前的占位实现。
type Requester interface {
	Publish(ctx context.Context, acc account.Account, art article.Article) (string, error)
}

// RepoCreator 在处理某个账号前视需要建一个"仓库/空间"。
type RepoCreator interface {
	CreateSpace(ctx context.Context, acc account.Account) error
}

// EventKind 标识一次发布过程中的事件类型。
type EventKind int

const (
	EventAttemptStart   EventKind = iota // 某账号的一次发布尝试开始
	EventAttemptSuccess                  // 该次尝试成功
	EventAttemptFailure                  // 该次尝试失败
	EventAccountSwitch                   // 放弃当前账号，换下一个（达到每号发布上限/连续失败换号/建仓库失败）
	EventRoundStart                      // 新的一轮开始
	EventRoundProgress                   // 本轮进度更新
	EventRoundDone                       // 本轮结束
)

// Event 是回传给上层的进度事件。
type Event struct {
	Kind         EventKind
	AccountIndex int // 账号在原始队列中的下标；EventRoundStart/RoundProgress/RoundDone 不适用
	CK           string
	ArticleTitle string
	Result       string // 成功时 Requester 返回的结果描述
	Err          error  // 失败/换号时的原因
	Round        int
	RoundTotal   int
	RoundDone    int
}

// IndexedAccount 携带账号在原始队列中的下标，用于事件回传时定位前端要更新的行。
type IndexedAccount struct {
	Index   int
	Account account.Account
}

// RunConfig 是一次批量发布任务的运行参数。
type RunConfig struct {
	Threads          int  // 并发线程数
	IntervalSec      int  // 同一账号相邻两次发布尝试之间的等待秒数
	PerAccountCount  int  // 单个账号最多发布多少次
	FailSwitchCount  int  // 账号连续失败达到此次数就换号
	CycleRounds      int  // 账号池整体循环轮数
	RoundIntervalSec int  // 相邻两轮之间的等待秒数
	CreateRepo       bool // 处理账号前是否先建仓库/空间
}

func (c *RunConfig) normalize() {
	if c.Threads < 1 {
		c.Threads = 1
	}
	if c.IntervalSec < 0 {
		c.IntervalSec = 0
	}
	if c.PerAccountCount < 1 {
		c.PerAccountCount = 1
	}
	if c.FailSwitchCount < 1 {
		c.FailSwitchCount = 1
	}
	if c.CycleRounds < 1 {
		c.CycleRounds = 1
	}
	if c.RoundIntervalSec < 0 {
		c.RoundIntervalSec = 0
	}
}

// PauseGate 是可在运行中途暂停/恢复批量任务的开关，多个 worker 共用同一个实例。
type PauseGate struct {
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

// NewPauseGate 创建一个初始为"未暂停"的开关。
func NewPauseGate() *PauseGate {
	g := &PauseGate{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

// Pause 暂停：worker 会在完成当前尝试后阻塞，直到 Resume 或 ctx 取消。
func (g *PauseGate) Pause() {
	g.mu.Lock()
	g.paused = true
	g.mu.Unlock()
}

// Resume 恢复运行。
func (g *PauseGate) Resume() {
	g.mu.Lock()
	g.paused = false
	g.mu.Unlock()
	g.cond.Broadcast()
}

// IsPaused 返回当前是否处于暂停状态。
func (g *PauseGate) IsPaused() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.paused
}

// wakeAll 唤醒所有等待者（不改变暂停状态），用于 ctx 取消时让 Wait 尽快返回。
func (g *PauseGate) wakeAll() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cond.Broadcast()
}

// wait 在暂停期间阻塞；ctx 取消时立即返回 ctx.Err()。
func (g *PauseGate) wait(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.paused {
		if err := ctx.Err(); err != nil {
			return err
		}
		g.cond.Wait()
	}
	return ctx.Err()
}

// Runner 并发执行账号发布任务。
type Runner struct {
	client Requester
	repo   RepoCreator
}

// New 创建 Runner；repo 为 nil 时忽略"创建仓库"选项。
func New(client Requester, repo RepoCreator) *Runner {
	return &Runner{client: client, repo: repo}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// Run 执行账号池的批量发布任务。gate 为 nil 时视为从不暂停。arts 为空时直接返回。
func (r *Runner) Run(ctx context.Context, cfg RunConfig, gate *PauseGate, pool []IndexedAccount, arts []article.Article, onEvent func(Event)) error {
	if len(arts) == 0 || len(pool) == 0 {
		return nil
	}
	cfg.normalize()
	if gate == nil {
		gate = NewPauseGate()
	}

	stopWake := make(chan struct{})
	defer close(stopWake)
	go func() {
		select {
		case <-ctx.Done():
			gate.wakeAll()
		case <-stopWake:
		}
	}()

	for round := 1; round <= cfg.CycleRounds; round++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		onEvent(Event{Kind: EventRoundStart, Round: round, RoundTotal: len(pool)})

		work := make(chan IndexedAccount, len(pool))
		for _, ia := range pool {
			work <- ia
		}
		close(work)

		var wg sync.WaitGroup
		var doneCount int32
		for t := 0; t < cfg.Threads; t++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for ia := range work {
					if ctx.Err() != nil {
						return
					}
					r.runAccount(ctx, cfg, gate, ia, arts, onEvent)
					n := atomic.AddInt32(&doneCount, 1)
					onEvent(Event{Kind: EventRoundProgress, Round: round, RoundDone: int(n), RoundTotal: len(pool)})
				}
			}()
		}
		wg.Wait()

		if err := ctx.Err(); err != nil {
			return err
		}
		onEvent(Event{Kind: EventRoundDone, Round: round, RoundTotal: len(pool)})

		if round < cfg.CycleRounds {
			if err := sleepCtx(ctx, time.Duration(cfg.RoundIntervalSec)*time.Second); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runner) runAccount(ctx context.Context, cfg RunConfig, gate *PauseGate, ia IndexedAccount, arts []article.Article, onEvent func(Event)) {
	if cfg.CreateRepo && r.repo != nil {
		if err := r.repo.CreateSpace(ctx, ia.Account); err != nil {
			onEvent(Event{Kind: EventAccountSwitch, AccountIndex: ia.Index, CK: ia.Account.CK, Err: err})
			return
		}
	}

	consecFail := 0
	for i := 0; i < cfg.PerAccountCount; i++ {
		if err := gate.wait(ctx); err != nil {
			return
		}
		if err := ctx.Err(); err != nil {
			return
		}

		art := arts[i%len(arts)]
		onEvent(Event{Kind: EventAttemptStart, AccountIndex: ia.Index, CK: ia.Account.CK, ArticleTitle: art.Title})

		result, err := r.client.Publish(ctx, ia.Account, art)
		if err != nil {
			consecFail++
			onEvent(Event{Kind: EventAttemptFailure, AccountIndex: ia.Index, CK: ia.Account.CK, ArticleTitle: art.Title, Err: err})
			if consecFail >= cfg.FailSwitchCount {
				onEvent(Event{Kind: EventAccountSwitch, AccountIndex: ia.Index, CK: ia.Account.CK, Err: errors.New("连续失败达到换号阈值")})
				return
			}
		} else {
			consecFail = 0
			onEvent(Event{Kind: EventAttemptSuccess, AccountIndex: ia.Index, CK: ia.Account.CK, ArticleTitle: art.Title, Result: result})
		}

		if i < cfg.PerAccountCount-1 && cfg.IntervalSec > 0 {
			if err := sleepCtx(ctx, time.Duration(cfg.IntervalSec)*time.Second); err != nil {
				return
			}
		}
	}
}

// TODORequester 是发布请求的占位实现：项目目前还没有接入目标系统的真实发布协议
// （CK 怎么带、UA/IP 怎么用、怎么判定成功失败），调用会直接返回错误，方便先跑通
// 账号队列的状态流转与累计统计。等接口细节确定后，实现一个新的 Requester 换掉它即可。
type TODORequester struct{}

func (TODORequester) Publish(ctx context.Context, acc account.Account, art article.Article) (string, error) {
	return "", errors.New("尚未接入目标系统的发布接口")
}

// TODORepoCreator 是"创建仓库/空间"的占位实现，逻辑同 TODORequester。
type TODORepoCreator struct{}

func (TODORepoCreator) CreateSpace(ctx context.Context, acc account.Account) error {
	return errors.New("尚未接入目标系统的建仓库接口")
}
