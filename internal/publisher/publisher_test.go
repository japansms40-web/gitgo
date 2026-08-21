package publisher

import (
	"context"
	"errors"
	"testing"

	"githubbaidu/internal/article"
	"githubbaidu/internal/config"
)

// fakeClient 实现 GitHubClient 接口，用于测试。
type fakeClient struct {
	putErr   map[string]error // repoPath -> 首次返回的错误
	putCalls map[string]int
}

func (f *fakeClient) GetFileSHA(ctx context.Context, owner, repo, path, branch string) (string, error) {
	return "", nil
}
func (f *fakeClient) PutFile(ctx context.Context, p PutParams) (string, error) {
	if f.putCalls == nil {
		f.putCalls = map[string]int{}
	}
	f.putCalls[p.Path]++
	if err, ok := f.putErr[p.Path]; ok && f.putCalls[p.Path] == 1 {
		return "", err // 只在首次失败，验证重试
	}
	return "https://example.com/" + p.Path, nil
}

func testCfg() config.Config {
	return config.Config{Owner: "o", Repo: "r", Branch: "main", Dir: "posts", IntervalSec: 0, Retries: 1}
}

func TestRun_AllSuccess(t *testing.T) {
	arts := []article.Article{
		{Title: "a", RepoPath: "posts/a.md", Content: []byte("x")},
		{Title: "b", RepoPath: "posts/b.md", Content: []byte("y")},
	}
	var events []Event
	p := New(&fakeClient{})
	err := p.Run(context.Background(), testCfg(), arts, func(e Event) { events = append(events, e) })
	if err != nil {
		t.Fatalf("Run err: %v", err)
	}
	// 每篇至少 Start + Success 两个事件
	var success int
	for _, e := range events {
		if e.Kind == EventSuccess {
			success++
			if e.URL == "" {
				t.Errorf("成功事件缺少 URL")
			}
		}
	}
	if success != 2 {
		t.Errorf("成功数 = %d, want 2", success)
	}
}

func TestRun_RetryThenSuccess(t *testing.T) {
	arts := []article.Article{{Title: "a", RepoPath: "posts/a.md", Content: []byte("x")}}
	fc := &fakeClient{putErr: map[string]error{"posts/a.md": errors.New("boom")}}
	var success, retry int
	p := New(fc)
	_ = p.Run(context.Background(), testCfg(), arts, func(e Event) {
		switch e.Kind {
		case EventSuccess:
			success++
		case EventRetry:
			retry++
		}
	})
	if success != 1 {
		t.Errorf("重试后应成功一次, success=%d", success)
	}
	if retry != 1 {
		t.Errorf("应有一次重试事件, retry=%d", retry)
	}
	if fc.putCalls["posts/a.md"] != 2 {
		t.Errorf("应调用 PutFile 两次, got %d", fc.putCalls["posts/a.md"])
	}
}

func TestRun_Cancel(t *testing.T) {
	arts := []article.Article{
		{Title: "a", RepoPath: "posts/a.md", Content: []byte("x")},
		{Title: "b", RepoPath: "posts/b.md", Content: []byte("y")},
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := New(&fakeClient{})
	var started int
	err := p.Run(ctx, testCfg(), arts, func(e Event) {
		if e.Kind == EventStart {
			started++
			cancel() // 第一篇开始后立即取消
		}
	})
	if err == nil {
		t.Errorf("取消后 Run 应返回 context 错误")
	}
	if started != 1 {
		t.Errorf("取消后不应开始第二篇, started=%d", started)
	}
}
