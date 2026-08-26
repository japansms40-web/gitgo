package publish

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"gitmd/internal/contentgen"
)

// captureReporter 收集回报，供断言。并发安全。
type captureReporter struct {
	mu       sync.Mutex
	logs     []string
	statuses map[int]string
	success  map[int]int
}

func newCapture() *captureReporter {
	return &captureReporter{statuses: map[int]string{}, success: map[int]int{}}
}
func (r *captureReporter) Log(kind, tag, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, tag+" "+msg)
}
func (r *captureReporter) Account(id int, status string, success, fail int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statuses[id] = status
	r.success[id] = success
}
func (r *captureReporter) Published(id int, repo, file string, urls ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range urls {
		r.logs = append(r.logs, "[链接] "+u)
	}
}

func liveProxy() string {
	for _, k := range []string{"GH_TEST_PROXY", "GITHUB_TEST_PROXY", "https_proxy", "HTTPS_PROXY", "http_proxy", "HTTP_PROXY"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// TestRunner_Live 端到端真跑发布引擎——【写操作】，双重开关守卫。
// 用最小素材库 + 真实账号，每号发 2 篇，验证 worker→验活→定仓→链式发文件全链路。
//
// 跑法：GH_TEST_COOKIE='...' GH_TEST_ALLOW_MUTATE=1 go test -run TestRunner_Live ./internal/publish/
func TestRunner_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络与有效 Cookie，-short 跳过")
	}
	if os.Getenv("GH_TEST_ALLOW_MUTATE") != "1" {
		t.Skip("写操作测试默认跳过：设 GH_TEST_ALLOW_MUTATE=1 才真发布")
	}
	ck := os.Getenv("GH_TEST_COOKIE")
	if ck == "" {
		t.Skip("设 GH_TEST_COOKIE 以跑发布集成测试")
	}

	lib := contentgen.Library{
		TitleTemplate: "claude-pub-{数字=6}",
		BodyTemplates: []string{"# claude 发布引擎测试\n随机 {数字=8}"},
	}

	rep := newCapture()
	// 2 轮 × 每号 1 篇：验证多轮语义（若轮次不生效，只会发 1 篇）。
	r := New(Config{
		Threads:          1,
		IntervalSec:      0,
		PerAccount:       1,
		FailSwitch:       3,
		Cycles:           2,
		RoundIntervalSec: 0,
		ProxyURL:         liveProxy(),
	}, []Account{{ID: 1, CK: ck}}, lib, contentgen.Options{}, rep, time.Now().UnixNano())

	r.Run(context.Background())

	if rep.statuses[1] != "success" {
		t.Fatalf("账号最终状态=%q, 期望 success；日志：%v", rep.statuses[1], rep.logs)
	}
	// 每轮 success 从 0 计，末轮应为 1；两轮共发 2 个文件（日志里两条 [发布]）。
	pubs := 0
	for _, l := range rep.logs {
		if len(l) >= 4 && l[:len("[发布]")] == "[发布]" {
			pubs++
		}
	}
	if pubs < 2 {
		t.Errorf("累计发布条数=%d, 期望 >= 2（2 轮各 1 篇）；日志：%v", pubs, rep.logs)
	}
	t.Logf("发布日志：%v", rep.logs)
}
