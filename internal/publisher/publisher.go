package publisher

import (
	"context"
	"time"

	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
)

// PutParams 复刻 github.PutParams 的字段，避免 publisher 直接依赖 github 包的具体类型，
// 便于测试注入 fake。字段含义与 github.PutParams 一致。
type PutParams struct {
	Owner, Repo, Path, Branch string
	Message                   string
	Content                   []byte
	SHA                       string
}

// GitHubClient 是 publisher 依赖的最小接口，由 github.Client 适配实现。
type GitHubClient interface {
	GetFileSHA(ctx context.Context, owner, repo, path, branch string) (string, error)
	PutFile(ctx context.Context, p PutParams) (string, error)
}

// EventKind 标识一次发布过程中的事件类型。
type EventKind int

const (
	EventStart   EventKind = iota // 开始发布某篇
	EventSuccess                  // 某篇成功
	EventFailure                  // 某篇最终失败
	EventRetry                    // 某篇失败后即将重试
)

// Event 是回传给 UI 的进度事件。
type Event struct {
	Kind  EventKind
	Index int // 文章在队列中的下标
	Title string
	URL   string // 成功时的 html_url
	Err   error  // 失败/重试时的原因
}

// Publisher 顺序发布文章。
type Publisher struct {
	client GitHubClient
}

// New 创建 Publisher。
func New(c GitHubClient) *Publisher {
	return &Publisher{client: c}
}

// Run 顺序发布所有文章，通过 onEvent 回传进度。ctx 取消则尽快停止并返回 ctx.Err()。
// 单篇失败不中断整批，仅记录 EventFailure。
func (p *Publisher) Run(ctx context.Context, cfg config.Config, arts []article.Article, onEvent func(Event)) error {
	cfg.Normalize()
	for i, a := range arts {
		if err := ctx.Err(); err != nil {
			return err
		}
		onEvent(Event{Kind: EventStart, Index: i, Title: a.Title})
		if err := ctx.Err(); err != nil { // 回调中可能已 cancel
			return err
		}

		url, err := p.publishOne(ctx, cfg, a, i, onEvent)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			onEvent(Event{Kind: EventFailure, Index: i, Title: a.Title, Err: err})
		} else {
			onEvent(Event{Kind: EventSuccess, Index: i, Title: a.Title, URL: url})
		}

		// 篇间间隔（最后一篇不等待），期间可被取消
		if i < len(arts)-1 && cfg.IntervalSec > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(cfg.IntervalSec) * time.Second):
			}
		}
	}
	return nil
}

func (p *Publisher) publishOne(ctx context.Context, cfg config.Config, a article.Article, idx int, onEvent func(Event)) (string, error) {
	var lastErr error
	// 尝试次数 = 1 + Retries
	for attempt := 0; attempt <= cfg.Retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if attempt > 0 {
			onEvent(Event{Kind: EventRetry, Index: idx, Title: a.Title, Err: lastErr})
		}
		sha, err := p.client.GetFileSHA(ctx, cfg.Owner, cfg.Repo, a.RepoPath, cfg.Branch)
		if err != nil {
			lastErr = err
			continue
		}
		url, err := p.client.PutFile(ctx, PutParams{
			Owner: cfg.Owner, Repo: cfg.Repo, Path: a.RepoPath, Branch: cfg.Branch,
			Message: "publish " + a.Title, Content: a.Content, SHA: sha,
		})
		if err != nil {
			lastErr = err
			continue
		}
		return url, nil
	}
	return "", lastErr
}
