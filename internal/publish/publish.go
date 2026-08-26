// Package publish 是批量发布的 worker-pool 引擎（适配自 InsExecutor 的任务/worker 模型）：
// 起 N 个 worker 并发消费账号，每账号独立处理——验活 + 看仓库数，为 0 则建仓，
// 再按内容配置往同一仓库连发多篇（链式提交）。进度经 Reporter 回报给上层（转 Wails 事件）。
package publish

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gitmd/internal/contentgen"
	"gitmd/internal/github"
)

// Account 是一个待发布账号。
type Account struct {
	ID int    // 前端账号 id，用于回报状态
	CK string // 整串会话 Cookie
}

// Config 是本轮批量发布生效的核心配置。
type Config struct {
	Threads     int    // 线程数量（并发 worker 数）
	IntervalSec int    // 每篇之间的间隔（秒）
	PerAccount  int    // 每号发布次数（往同一仓库塞的文件数）
	FailSwitch  int    // 累计失败达到此数就换号（停掉该账号剩余发布）
	ProxyURL    string // 代理（启用时非空，空=直连）
}

// Reporter 把进度回报给上层。实现必须并发安全（多 worker 并发调用）。
type Reporter interface {
	Log(kind, tag, msg string)                        // 一行日志
	Account(id int, status string, success, fail int) // 账号状态更新
}

// Runner 执行一次批量发布。
type Runner struct {
	cfg      Config
	accounts []Account
	lib      contentgen.Library
	genOpts  contentgen.Options
	report   Reporter
	seed     int64
}

// New 构造 Runner，并把非法配置纠正到可用下限。
func New(cfg Config, accounts []Account, lib contentgen.Library, genOpts contentgen.Options, report Reporter, seed int64) *Runner {
	if cfg.Threads < 1 {
		cfg.Threads = 1
	}
	if cfg.PerAccount < 1 {
		cfg.PerAccount = 1
	}
	if cfg.FailSwitch < 1 {
		cfg.FailSwitch = 1
	}
	return &Runner{cfg: cfg, accounts: accounts, lib: lib, genOpts: genOpts, report: report, seed: seed}
}

// Run 起 N 个 worker 并发消费账号，直到全部处理完或 ctx 取消。阻塞至结束。
func (r *Runner) Run(ctx context.Context) {
	ch := make(chan Account)
	var wg sync.WaitGroup
	for w := 0; w < r.cfg.Threads; w++ {
		wg.Add(1)
		go r.worker(ctx, w, ch, &wg)
	}
	// 投递账号；ctx 取消或投递完即关闭 channel，worker range 自然退出。
	go func() {
		defer close(ch)
		for _, a := range r.accounts {
			select {
			case <-ctx.Done():
				return
			case ch <- a:
			}
		}
	}()
	wg.Wait()
}

// worker 从 channel 取账号处理，每 worker 一个独立随机源（避免共享 rnd 竞争）。
func (r *Runner) worker(ctx context.Context, id int, ch <-chan Account, wg *sync.WaitGroup) {
	defer wg.Done()
	rnd := rand.New(rand.NewSource(r.seed + int64(id)*1_000_003 + 1))
	for a := range ch {
		select {
		case <-ctx.Done():
			return
		default:
		}
		r.processAccount(ctx, a, rnd)
	}
}

// processAccount 处理单个账号：验活 → 定目标仓库（为 0 建仓）→ 连发 PerAccount 篇。
func (r *Runner) processAccount(ctx context.Context, a Account, rnd *rand.Rand) {
	r.report.Account(a.ID, "publishing", 0, 0)

	client, err := github.New(a.CK, github.WithProxy(r.cfg.ProxyURL))
	if err != nil {
		r.fail(a, "构造客户端失败："+err.Error())
		return
	}

	// 1. 验活 + 仓库数
	repos, err := client.ListRepos(ctx, 1)
	if err != nil {
		if code := client.LastStatusCode(); code == 401 || code == 403 {
			r.report.Account(a.ID, "bad", 0, 0)
			r.report.Log("failure", "[坏号]", fmt.Sprintf("%s · HTTP %d", short(a.CK), code))
			return
		}
		r.fail(a, "验活失败："+err.Error())
		return
	}
	route := repos.Payload.ReposFinderPageRoute

	owner := github.OwnerFromCookie(a.CK)
	if owner == "" && len(route.Repositories) > 0 {
		owner = route.Repositories[0].Owner
	}
	if owner == "" {
		r.fail(a, "无法确定 owner（Cookie 缺 dotcom_user）")
		return
	}

	// 内容：一次性生成 PerAccount 篇（让顺序关键词等能在批内轮转）。
	drafts := r.genBatch(rnd, r.cfg.PerAccount)
	if len(drafts) == 0 {
		r.fail(a, "生成内容为空，请先在「内容设置」配置模板/词库")
		return
	}

	// 2. 定目标仓库
	repoName := ""
	baseCommit := ""
	if route.RepositoryCount == 0 || len(route.Repositories) == 0 {
		repoName = sanitizeRepoName(drafts[0].Title)
		if _, err := client.CreateRepo(ctx, github.CreateRepoParams{Owner: owner, Name: repoName, Visibility: "public"}); err != nil {
			r.fail(a, "建仓库失败："+err.Error())
			return
		}
		r.report.Log("info", "[建仓]", fmt.Sprintf("%s/%s", owner, repoName))
		baseCommit = "" // 新空仓库首次提交无父 commit
	} else {
		repoName = route.Repositories[0].Name
		// 取已有仓库当前 HEAD 作首篇父提交；失败则留空由服务端兜底。
		if ic, err := client.GetImplicitContext(ctx, owner, repoName, "/"+owner+"/"+repoName+"/new/main"); err == nil {
			baseCommit = ic.CommitOID
		}
	}

	// 3. 连发 PerAccount 篇（链式提交），累计失败达 FailSwitch 换号。
	success, fail := 0, 0
	for i, d := range drafts {
		select {
		case <-ctx.Done():
			r.report.Account(a.ID, finalStatus(success), success, fail)
			return
		default:
		}

		resp, ferr := client.CreateFile(ctx, github.CreateFileParams{
			Owner:      owner,
			Repo:       repoName,
			Branch:     "main",
			Filename:   sanitizeFilename(d.Title),
			Content:    d.Body,
			Message:    firstLine(d.Title),
			BaseCommit: baseCommit,
		})
		if ferr != nil {
			fail++
			r.report.Account(a.ID, "publishing", success, fail)
			r.report.Log("failure", "[失败]", fmt.Sprintf("%s · %v", short(a.CK), ferr))
			if fail >= r.cfg.FailSwitch {
				r.report.Log("retry", "[换号]", fmt.Sprintf("%s 累计失败 %d，停发", short(a.CK), fail))
				break
			}
			continue
		}
		success++
		baseCommit = commitSHAFromQuorumPath(resp.Data.CommitQuorumPollPath)
		r.report.Account(a.ID, "publishing", success, fail)
		r.report.Log("success", "[发布]", fmt.Sprintf("%s/%s ← %s", owner, repoName, sanitizeFilename(d.Title)))

		if i < len(drafts)-1 && !sleepCtx(ctx, time.Duration(r.cfg.IntervalSec)*time.Second) {
			break // 被取消
		}
	}

	r.report.Account(a.ID, finalStatus(success), success, fail)
}

// genBatch 生成 n 篇草稿；模板/词库为空或出错时返回 nil，由调用方判空。
func (r *Runner) genBatch(rnd *rand.Rand, n int) []contentgen.Draft {
	opts := r.genOpts
	opts.Count = n
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	drafts, err := contentgen.Generate(r.lib, opts, rnd, now)
	if err != nil {
		return nil
	}
	return drafts
}

// fail 统一回报账号失败（非坏号，如网络/代理/建仓/内容问题）。
func (r *Runner) fail(a Account, msg string) {
	r.report.Account(a.ID, "failed", 0, 0)
	r.report.Log("failure", "[失败]", fmt.Sprintf("%s · %s", short(a.CK), msg))
}

// finalStatus 按成功篇数决定账号最终状态。
func finalStatus(success int) string {
	if success > 0 {
		return "success"
	}
	return "failed"
}

// sleepCtx 睡 d，期间被取消则提前返回 false。
func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// short 截断 CK 便于日志展示。
func short(ck string) string {
	if len(ck) > 16 {
		return ck[:16] + "…"
	}
	return ck
}

// firstLine 取字符串首行（作提交信息，避免整段正文当 commit message）。
func firstLine(s string) string {
	if i := indexNewline(s); i >= 0 {
		return s[:i]
	}
	return s
}

func indexNewline(s string) int {
	for i, r := range s {
		if r == '\n' || r == '\r' {
			return i
		}
	}
	return -1
}
