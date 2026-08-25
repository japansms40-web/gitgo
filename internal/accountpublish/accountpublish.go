// Package accountpublish 把 dfclientkit/taskrunner 这个通用并发引擎实例化到
// article.Article 这个具体的处理对象类型上，是本项目专属的薄封装层。
package accountpublish

import (
	"context"
	"errors"

	"dfclientkit/account"
	"dfclientkit/taskrunner"

	"githubbaidu/internal/article"
)

type (
	EventKind      = taskrunner.EventKind
	Event          = taskrunner.Event
	IndexedAccount = taskrunner.IndexedAccount
	RunConfig      = taskrunner.RunConfig
	PauseGate      = taskrunner.PauseGate
	RepoCreator    = taskrunner.RepoCreator
	Requester      = taskrunner.Requester[article.Article]
)

const (
	EventAttemptStart   = taskrunner.EventAttemptStart
	EventAttemptSuccess = taskrunner.EventAttemptSuccess
	EventAttemptFailure = taskrunner.EventAttemptFailure
	EventAccountSwitch  = taskrunner.EventAccountSwitch
	EventRoundStart     = taskrunner.EventRoundStart
	EventRoundProgress  = taskrunner.EventRoundProgress
	EventRoundDone      = taskrunner.EventRoundDone
)

var NewPauseGate = taskrunner.NewPauseGate

func itemLabel(a article.Article) string { return a.Title }

// Runner 并发执行"用账号队列处理文章"的任务；内部把文章的标题喂给 taskrunner
// 当作事件展示文案。
type Runner struct {
	inner *taskrunner.Runner[article.Article]
}

// New 创建 Runner；repo 为 nil 时忽略"创建仓库"选项。
func New(client Requester, repo RepoCreator) *Runner {
	return &Runner{inner: taskrunner.New[article.Article](client, repo)}
}

// Run 执行账号池的批量发布任务，签名与迁移前保持一致。
func (r *Runner) Run(ctx context.Context, cfg RunConfig, gate *PauseGate, pool []IndexedAccount, arts []article.Article, onEvent func(Event)) error {
	return r.inner.Run(ctx, cfg, gate, pool, arts, itemLabel, onEvent)
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
