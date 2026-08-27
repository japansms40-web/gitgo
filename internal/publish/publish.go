// Package publish 是批量发布的 worker-pool 引擎（适配自 InsExecutor 的任务/worker 模型）：
// 起 N 个 worker 并发消费账号，每账号独立处理——验活 + 看仓库数，为 0 则建仓，
// 再按内容配置往同一仓库连发多篇（链式提交）。进度经 Reporter 回报给上层（转 Wails 事件）。
package publish

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
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
	Threads          int    // 线程数量（并发 worker 数）
	IntervalSec      int    // 每篇之间的间隔（秒）
	PerAccount       int    // 每号发布次数（往同一仓库塞的文件数）
	FailSwitch       int    // 累计失败达到此数就换号（停掉该账号剩余发布）
	Cycles           int    // 账号循环轮数（整个账号列表跑几遍）
	RoundIntervalSec int    // 每轮之间的间隔（秒）
	ProxyURL         string // 代理（启用时非空，空=直连）
}

// Reporter 把进度回报给上层。实现必须并发安全（多 worker 并发调用）。
type Reporter interface {
	Log(kind, tag, msg string)                           // 一行日志
	Account(id int, status string, success, fail int)    // 账号状态更新
	Published(id int, repo, file string, urls ...string) // 成功发布一篇（一篇多条 GitHub 链接，须成对原子落盘）
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
	if cfg.Cycles < 1 {
		cfg.Cycles = 1
	}
	return &Runner{cfg: cfg, accounts: accounts, lib: lib, genOpts: genOpts, report: report, seed: seed}
}

// Run 按「账号循环」跑多轮：每轮起 N 个 worker 并发消费一遍全部账号，轮间等待 RoundIntervalSec。
// ctx 取消随时提前退出。阻塞至结束。
func (r *Runner) Run(ctx context.Context) {
	for round := 1; round <= r.cfg.Cycles; round++ {
		if ctx.Err() != nil {
			return
		}
		if r.cfg.Cycles > 1 {
			r.report.Log("start", "[轮次]", fmt.Sprintf("第 %d/%d 轮开始", round, r.cfg.Cycles))
		}
		r.runRound(ctx, round)
		if ctx.Err() != nil {
			return
		}
		if round < r.cfg.Cycles {
			r.report.Log("info", "[轮次]", fmt.Sprintf("第 %d 轮结束，等待 %d 秒后下一轮", round, r.cfg.RoundIntervalSec))
			if !sleepCtx(ctx, time.Duration(r.cfg.RoundIntervalSec)*time.Second) {
				return
			}
		}
	}
}

// runRound 跑一轮：起 N 个 worker 并发消费全部账号，等这一轮收尾。
func (r *Runner) runRound(ctx context.Context, round int) {
	ch := make(chan Account)
	var wg sync.WaitGroup
	for w := 0; w < r.cfg.Threads; w++ {
		wg.Add(1)
		go r.worker(ctx, round, w, ch, &wg)
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

// worker 从 channel 取账号处理，每 (轮次,worker) 一个独立随机源：
// 加 round 让每轮生成的内容不同，加 id 避免同轮 worker 间共享 rnd 竞争。
func (r *Runner) worker(ctx context.Context, round, id int, ch <-chan Account, wg *sync.WaitGroup) {
	defer wg.Done()
	rnd := rand.New(rand.NewSource(r.seed + int64(round)*1_000_000_007 + int64(id)*1_000_003 + 1))
	for a := range ch {
		select {
		case <-ctx.Done():
			return
		default:
		}
		r.processAccount(ctx, a, rnd)
	}
}

// log 是 report.Log 的格式化便捷封装。kind 决定前端颜色（info/success/failure/retry/debug）。
func (r *Runner) log(kind, tag, format string, args ...any) {
	r.report.Log(kind, tag, fmt.Sprintf(format, args...))
}

// processAccount 处理单个账号：验活 → 定目标仓库（为 0 建仓）→ 连发 PerAccount 篇。
// 全程按 请求/响应/调试/错误 打点，日志落盘后便于事后（含 AI）排查每一步的请求与返回。
func (r *Runner) processAccount(ctx context.Context, a Account, rnd *rand.Rand) {
	ck := fmt.Sprintf("账号#%d", a.ID) // 日志标签用账号序号，避免把 Cookie 片段写进磁盘日志
	r.report.Account(a.ID, "publishing", 0, 0)
	r.log("start", "[账号]", "开始处理 %s", ck)

	proxyDesc := r.cfg.ProxyURL
	if proxyDesc == "" {
		proxyDesc = "直连"
	}
	r.log("debug", "[调试]", "%s 代理=%s", ck, proxyDesc)

	client, err := github.New(a.CK, github.WithProxy(r.cfg.ProxyURL))
	if err != nil {
		r.fail(a, "构造客户端失败："+err.Error())
		return
	}

	// 1. 验活 + 仓库数
	r.log("debug", "[请求]", "%s GET /repos?q=owner:@me&page=1", ck)
	repos, err := client.ListRepos(ctx, 1)
	code := client.LastStatusCode()
	if err != nil {
		if code == 401 || code == 403 {
			r.report.Account(a.ID, "bad", 0, 0)
			r.log("failure", "[坏号]", "%s 验活 HTTP %d（鉴权失败）", ck, code)
			return
		}
		r.report.Account(a.ID, "failed", 0, 0)
		r.log("failure", "[错误]", "%s 验活失败 HTTP %d：%v", ck, code, err)
		return
	}
	route := repos.Payload.ReposFinderPageRoute
	r.log("info", "[响应]", "%s 验活 HTTP %d · 仓库 %d 个", ck, code, route.RepositoryCount)

	owner := github.OwnerFromCookie(a.CK)
	if owner == "" && len(route.Repositories) > 0 {
		owner = route.Repositories[0].Owner
	}
	if owner == "" {
		r.fail(a, "无法确定 owner（Cookie 缺 dotcom_user）")
		return
	}
	r.log("debug", "[调试]", "%s owner=%s", ck, owner)

	// 内容：一次性生成 PerAccount 篇（让顺序关键词等能在批内轮转）。
	drafts := r.genBatch(rnd, r.cfg.PerAccount)
	if len(drafts) == 0 {
		r.fail(a, "生成内容为空，请先在「内容设置」配置模板/词库")
		return
	}
	r.log("debug", "[调试]", "%s 生成 %d 篇待发", ck, len(drafts))

	// 2. 定目标仓库
	repoName := ""
	baseCommit := ""
	if route.RepositoryCount == 0 || len(route.Repositories) == 0 {
		repoName = sanitizeRepoName(drafts[0].Title)
		r.log("debug", "[请求]", "%s POST /repositories 建仓 %s/%s", ck, owner, repoName)
		if _, err := client.CreateRepo(ctx, github.CreateRepoParams{Owner: owner, Name: repoName, Visibility: "public"}); err != nil {
			r.report.Account(a.ID, "failed", 0, 0)
			r.log("failure", "[错误]", "%s 建仓失败 HTTP %d：%v", ck, client.LastStatusCode(), err)
			return
		}
		r.log("success", "[建仓]", "%s/%s（HTTP %d）", owner, repoName, client.LastStatusCode())
		baseCommit = "" // 新空仓库首次提交无父 commit
	} else {
		repoName = route.Repositories[0].Name
		r.log("info", "[信息]", "%s 用已有仓库 %s/%s", ck, owner, repoName)
		r.log("debug", "[请求]", "%s GET /github-copilot/chat/implicit-context/%s/%s/... 取父提交", ck, owner, repoName)
		if ic, err := client.GetImplicitContext(ctx, owner, repoName, "/"+owner+"/"+repoName+"/new/main"); err == nil {
			baseCommit = ic.CommitOID
			r.log("debug", "[响应]", "%s 父提交=%s（HTTP %d）", ck, baseCommit, client.LastStatusCode())
		} else {
			r.log("retry", "[调试]", "%s 取父提交失败，留空由服务端兜底：%v", ck, err)
		}
	}

	// 3. 连发 PerAccount 篇（链式提交），累计失败达 FailSwitch 换号。
	success, fail := 0, 0
	for i, d := range drafts {
		select {
		case <-ctx.Done():
			r.log("retry", "[停止]", "%s 被取消，已发 %d 篇", ck, success)
			r.report.Account(a.ID, finalStatus(success), success, fail)
			return
		default:
		}

		fn := sanitizeFilename(d.Title)
		r.log("debug", "[请求]", "%s POST /%s/%s/create/main ← %s（第 %d/%d 篇）", ck, owner, repoName, fn, i+1, len(drafts))
		resp, ferr := client.CreateFile(ctx, github.CreateFileParams{
			Owner:      owner,
			Repo:       repoName,
			Branch:     "main",
			Filename:   fn,
			Content:    d.Body,
			Message:    firstLine(d.Title),
			// 正文同时塞进 commit 说明正文：GitHub commit 页会把它显示在标题下方，
			// 并把其中的 http(s):// 与 www. 域名自动渲染成可点链接（与同行页面一致）。
			Description: d.Body,
			BaseCommit:  baseCommit,
		})
		fcode := client.LastStatusCode()
		if ferr != nil {
			fail++
			r.report.Account(a.ID, "publishing", success, fail)
			r.log("failure", "[错误]", "%s 发文件失败 HTTP %d：%v", ck, fcode, ferr)
			if fail >= r.cfg.FailSwitch {
				r.log("retry", "[换号]", "%s 累计失败 %d，停发换号", ck, fail)
				break
			}
			continue
		}
		success++
		baseCommit = commitSHAFromQuorumPath(resp.Data.CommitQuorumPollPath)
		// 每篇输出两条链接：commit（提交）+ blob（文件），成对上报、原子落盘。
		commitURL := "https://github.com/" + owner + "/" + repoName + "/commit/" + baseCommit
		blobURL := "https://github.com/" + owner + "/" + repoName + "/blob/main/" + url.PathEscape(fn)
		r.report.Account(a.ID, "publishing", success, fail)
		r.report.Published(a.ID, repoName, fn, commitURL, blobURL)
		r.log("info", "[响应]", "%s HTTP %d commit=%s", ck, fcode, baseCommit)
		r.log("success", "[发布]", "%s/%s ← %s", owner, repoName, fn)

		if i < len(drafts)-1 && !sleepCtx(ctx, time.Duration(r.cfg.IntervalSec)*time.Second) {
			r.log("retry", "[停止]", "%s 被取消", ck)
			break
		}
	}

	r.log("info", "[账号]", "%s 完成：成功 %d 失败 %d", ck, success, fail)
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
	r.log("failure", "[错误]", "账号#%d · %s", a.ID, msg)
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
