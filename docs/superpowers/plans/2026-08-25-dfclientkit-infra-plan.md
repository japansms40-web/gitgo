# dfclientkit 共用基建 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 githubbaidu（`ghpublisher`）里已经足够通用的账号队列、并发发布引擎、配置持久化、日志格式化、桌面壳 UI 抽成独立仓库 `dfclientkit`，githubbaidu 改为消费它，作为以后新平台发布器项目的基建参照实现。

**Architecture:** 新建独立仓库 `dfclientkit`，内含一个零外部依赖的 Go 模块（`account`/`appconfig`/`taskrunner`/`runlog` 四个包）和一个 source-only 的 npm 包 `@dongfang/df-ui-shell`（六个 Vue3 组件 + 基础主题）。githubbaidu 通过 `go.mod` 的本地 `replace` 和 `package.json` 的 `file:` 依赖消费它；githubbaidu 内的 `internal/account`、`internal/config`、`internal/accountpublish`、`internal/runlog` 改为薄壳委托层，`app.go` 的对外方法签名基本不变。

**Tech Stack:** Go 1.26（含泛型）、Vue 3 + Vite（`@vitejs/plugin-vue`）、Wails v2。

**参照设计文档：** `docs/superpowers/specs/2026-08-25-dfclientkit-infra-design.md`

---

## 关于本计划的两点说明

1. **本计划里 Go 侧的每一段代码都已经在 scratchpad 里实际跑过 `go build`/`go vet`/`go test`，全部通过**（包括泛型相关的 `taskrunner.Runner[Item]`、`appconfig.Load[T]`，以及 githubbaidu 侧薄壳委托层如何调用它们）；Vue 组件也都用项目里已装好的 `vue/compiler-sfc` 做过语法编译校验。所以下面各 Task 里"写测试 → 跑测试失败 → 写实现 → 跑测试通过"这几步，你只是在把已经验证过的代码誊到正式仓库里，不会遇到设计上的意外。
2. **前端组件迁移没有走 TDD 的"先写测试"套路**——这个项目本身没有给 Vue 组件配单元测试框架（没有 vitest/jest），迁移前的六个组件也都没有测试。所以 Phase C（ui-shell）和 Phase E（githubbaidu 前端接入）里的步骤是"创建/修改文件 → 提交"，真正的验证放在最后 Task 22 用 `npm run build` + `wails dev` 手动过一遍。

---

## Phase A — 初始化 dfclientkit 仓库

### Task 1: 创建 dfclientkit 仓库骨架

**Files:**
- Create: `~/GolandProjects/dfclientkit/go/go.mod`
- Create: `~/GolandProjects/dfclientkit/README.md`
- Create: `~/GolandProjects/dfclientkit/.gitignore`

- [ ] **Step 1: 创建目录结构**

Run:
```bash
mkdir -p ~/GolandProjects/dfclientkit/go/{account,appconfig,taskrunner,runlog}
mkdir -p ~/GolandProjects/dfclientkit/ui-shell/src
```

- [ ] **Step 2: 写 go.mod**

`~/GolandProjects/dfclientkit/go/go.mod`:
```
module dfclientkit

go 1.26
```

- [ ] **Step 3: 写 .gitignore**

`~/GolandProjects/dfclientkit/.gitignore`:
```
.DS_Store
```

- [ ] **Step 4: 写 README**

`~/GolandProjects/dfclientkit/README.md`:
```markdown
# dfclientkit

多个"账号队列驱动"型桌面客户端项目共用的基建。每个消费方独立打包、独立运行，
这里只提供可复用的底层能力。

## 目录

- `go/` — Go 模块（`module dfclientkit`），四个包：
  - `account` — 账号队列模型（CK/UA/IP + 状态统计）与磁盘持久化
  - `appconfig` — 通用的 JSON 配置读写 + 系统默认目录解析
  - `taskrunner` — 账号队列驱动的并发任务引擎（换号/多轮/暂停恢复/取消）
  - `runlog` — 把 taskrunner 的事件格式化成运行日志文案
- `ui-shell/` — npm 包 `@dongfang/df-ui-shell`：桌面壳的 Vue3 组件（标题栏/侧边导航/
  状态栏/日志面板/数字输入/结果弹窗）与基础主题，纯源码分发，不预编译。

## 本地开发中如何被消费

消费方项目在同级目录下（比如 `~/GolandProjects/<consumer>`）：

`go.mod`：
```
require dfclientkit v0.0.0
replace dfclientkit => ../dfclientkit/go
```

`frontend/package.json`：
```json
"@dongfang/df-ui-shell": "file:../../dfclientkit/ui-shell"
```
```

- [ ] **Step 5: git init 并提交**

Run:
```bash
cd ~/GolandProjects/dfclientkit
git init
git add go/go.mod README.md .gitignore
git commit -m "chore: 初始化 dfclientkit 仓库骨架"
```

Expected: 提交成功，`git log --oneline` 能看到这一条记录。

---

### Task 2: account 包

**Files:**
- Create: `~/GolandProjects/dfclientkit/go/account/account.go`
- Test: `~/GolandProjects/dfclientkit/go/account/account_test.go`

- [ ] **Step 1: 写测试文件**

`~/GolandProjects/dfclientkit/go/account/account_test.go`:
```go
package account

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseImportText(t *testing.T) {
	text := "  ck-one \n----\nck-two\n----\n   \n----\nck-three"
	got := ParseImportText(text)
	want := []string{"ck-one", "ck-two", "ck-three"}
	if len(got) != len(want) {
		t.Fatalf("解析出 %d 个账号, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].CK != w {
			t.Errorf("got[%d].CK = %q, want %q", i, got[i].CK, w)
		}
		if got[i].Status != StatusPending {
			t.Errorf("got[%d].Status = %q, want %q", i, got[i].Status, StatusPending)
		}
	}
}

func TestParseImportTextEmpty(t *testing.T) {
	if got := ParseImportText("   \n----\n\n"); len(got) != 0 {
		t.Errorf("空文本应解析出 0 个账号, got %+v", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "accounts.json")
	list := []Account{
		{CK: "a", UA: "Chrome", IP: "1.2.3.4", Status: StatusSuccess, Success: 3, Total: 3},
		{CK: "b", Status: StatusFailed, Fail: 1, Total: 1, Bad: true},
	}
	if err := Save(path, list); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got := Load(path)
	if len(got) != len(list) {
		t.Fatalf("Load() = %+v, want %+v", got, list)
	}
	for i := range list {
		if got[i] != list[i] {
			t.Errorf("Load()[%d] = %+v, want %+v", i, got[i], list[i])
		}
	}
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	if got := Load(path); len(got) != 0 {
		t.Errorf("Load() = %+v, want empty", got)
	}
}

// 回归测试：Load() 必须返回非 nil 的切片，序列化成 JSON 才是 "[]" 而不是
// "null"。以前这里返回了裸 nil，消费方前端 accounts.value 被赋成 null，一读
// accounts.length 整个页面就崩，点哪都没反应。
func TestLoadMissingFileMarshalsToEmptyArrayNotNull(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	b, err := json.Marshal(Load(path))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(b) != "[]" {
		t.Errorf("Load() 序列化结果 = %s, want []（不能是 null）", b)
	}
}

// 回归测试：磁盘上的文件如果是字面量 "null"，重新读出来也必须变回空数组，
// 不能继续把 null 传染下去。
func TestLoadLiteralNullJSONMarshalsToEmptyArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	if err := os.WriteFile(path, []byte("null"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	b, err := json.Marshal(Load(path))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(b) != "[]" {
		t.Errorf("Load() 序列化结果 = %s, want []（不能是 null）", b)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./account/...`
Expected: FAIL，编译错误 `undefined: ParseImportText` / `undefined: Account` / `undefined: StatusPending` 等（因为 `account.go` 还不存在）。

- [ ] **Step 3: 写实现**

`~/GolandProjects/dfclientkit/go/account/account.go`:
```go
// Package account 管理"账号队列驱动"型客户端项目共用的账号模型：CK/UA/IP 三元组、
// 状态与累计统计、导入文本解析、磁盘持久化。不依赖任何 UI 或具体发布协议。
package account

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// 账号状态取值；与前端展示的中文文案保持一致。
const (
	StatusPending = "待发"
	StatusRunning = "发布中"
	StatusSuccess = "成功"
	StatusFailed  = "失败"
)

// Account 是队列中的一个账号。json tag 同时决定磁盘持久化格式与前端绑定收到的字段名。
type Account struct {
	CK      string `json:"ck"`
	UA      string `json:"ua"`
	IP      string `json:"ip"`
	Status  string `json:"status"`
	Success int    `json:"success"`
	Fail    int    `json:"fail"`
	Total   int    `json:"total"`
	Bad     bool   `json:"bad"` // 手动标记为坏号，批量处理时跳过
}

// ParseImportText 把粘贴/拖入的文本按 "----" 分隔解析为账号列表；
// 每段去除首尾空白后作为一个账号的 CK，空段落忽略。
func ParseImportText(text string) []Account {
	parts := strings.Split(text, "----")
	out := make([]Account, 0, len(parts))
	for _, p := range parts {
		ck := strings.TrimSpace(p)
		if ck == "" {
			continue
		}
		out = append(out, Account{CK: ck, Status: StatusPending})
	}
	return out
}

// Load 从 path 指向的 JSON 文件读取账号列表；文件不存在或内容损坏时返回空列表。
// 返回值始终是非 nil 的切片：Go 的 nil 切片编码成 JSON 会变成 null，传到前端会让
// `accounts.value` 变成 null 而不是空数组，导致模板里任何 accounts.length /
// accounts.map 之类的访问直接抛异常，把整个页面渲染搞崩——这里必须显式兜底，
// 不能只判断 len() 是否为 0。
func Load(path string) []Account {
	list := []Account{}
	b, err := os.ReadFile(path)
	if err != nil {
		return list
	}
	_ = json.Unmarshal(b, &list)
	if list == nil {
		list = []Account{}
	}
	return list
}

// Save 把账号列表写入 path 指向的 JSON 文件，按需创建父目录。
func Save(path string, list []Account) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./account/... -v`
Expected: PASS，7 个测试全绿。

- [ ] **Step 5: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add go/account
git commit -m "feat: account 包——账号队列模型与持久化"
```

---

### Task 3: appconfig 包

**Files:**
- Create: `~/GolandProjects/dfclientkit/go/appconfig/appconfig.go`
- Test: `~/GolandProjects/dfclientkit/go/appconfig/appconfig_test.go`

- [ ] **Step 1: 写测试文件**

`~/GolandProjects/dfclientkit/go/appconfig/appconfig_test.go`:
```go
package appconfig

import (
	"path/filepath"
	"testing"
)

type testConfig struct {
	Threads int    `json:"threads"`
	Name    string `json:"name"`
}

func TestLoadFileSaveFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	cfg := testConfig{Threads: 5, Name: "hi"}
	if err := SaveFile(path, cfg); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}
	got := LoadFile(path, testConfig{})
	if got != cfg {
		t.Errorf("LoadFile() = %+v, want %+v", got, cfg)
	}
}

func TestLoadFileMissingReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	defaults := testConfig{Threads: 1, Name: "default"}
	if got := LoadFile(path, defaults); got != defaults {
		t.Errorf("LoadFile() = %+v, want %+v", got, defaults)
	}
}

func TestLoadFileCorruptJSONReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := SaveFile(path, "not-json-shaped-but-valid-string"); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}
	defaults := testConfig{Threads: 9, Name: "fallback"}
	got := LoadFile(path, defaults)
	if got != defaults {
		t.Errorf("LoadFile() = %+v, want %+v（JSON 字符串不是合法的 testConfig 形状应回退默认值）", got, defaults)
	}
}

func TestDefaultDirIncludesAppName(t *testing.T) {
	dir, err := DefaultDir("ghpublisher")
	if err != nil {
		t.Fatalf("DefaultDir() error = %v", err)
	}
	if filepath.Base(dir) != "ghpublisher" {
		t.Errorf("DefaultDir() = %q, want basename ghpublisher", dir)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./appconfig/...`
Expected: FAIL，编译错误 `undefined: SaveFile` / `undefined: LoadFile` / `undefined: DefaultDir`。

- [ ] **Step 3: 写实现**

`~/GolandProjects/dfclientkit/go/appconfig/appconfig.go`:
```go
// Package appconfig 提供"读 JSON 配置、按系统目录存、缺省兜底"这套通用持久化机制，
// 供各客户端项目的运行参数结构体复用。
package appconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultDir 返回 appName 对应应用在当前系统上的默认状态目录
// （macOS: ~/Library/Application Support/<appName>，Windows: %AppData%/<appName>）。
func DefaultDir(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName), nil
}

// LoadFile 从 path 指向的 JSON 文件读取配置；文件不存在或内容无法解析时返回 defaults。
func LoadFile[T any](path string, defaults T) T {
	b, err := os.ReadFile(path)
	if err != nil {
		return defaults
	}
	cfg := defaults
	if err := json.Unmarshal(b, &cfg); err != nil {
		return defaults
	}
	return cfg
}

// SaveFile 把 cfg 写入 path 指向的 JSON 文件，按需创建父目录。
func SaveFile[T any](path string, cfg T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Load 从 appName 对应默认目录下的 config.json 读取配置；目录/文件不存在或内容
// 无法解析时返回 defaults。
func Load[T any](appName string, defaults T) T {
	dir, err := DefaultDir(appName)
	if err != nil {
		return defaults
	}
	return LoadFile(filepath.Join(dir, "config.json"), defaults)
}

// Save 把 cfg 写入 appName 对应默认目录下的 config.json，按需创建父目录。
func Save[T any](appName string, cfg T) error {
	dir, err := DefaultDir(appName)
	if err != nil {
		return err
	}
	return SaveFile(filepath.Join(dir, "config.json"), cfg)
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./appconfig/... -v`
Expected: PASS，4 个测试全绿。

- [ ] **Step 5: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add go/appconfig
git commit -m "feat: appconfig 包——通用 JSON 配置持久化"
```

---

### Task 4: taskrunner 包

这是最大的一块：把 githubbaidu 里 `internal/accountpublish` 的并发发布引擎整段搬过来，做两处改动——包名去掉发布/GitHub 语境，`Requester`/`Runner` 用 Go 泛型参数化"被处理对象的类型"（`Item`），`Event.ArticleTitle` 改名 `Event.ItemLabel`，由调用方传入的 `itemLabel func(Item) string` 提取。`RunConfig`（含 `CreateRepo`）、`PauseGate`、并发/换号/多轮/暂停恢复/取消的逻辑本身一字不改。

**Files:**
- Create: `~/GolandProjects/dfclientkit/go/taskrunner/taskrunner.go`
- Test: `~/GolandProjects/dfclientkit/go/taskrunner/taskrunner_test.go`

- [ ] **Step 1: 写测试文件**

`~/GolandProjects/dfclientkit/go/taskrunner/taskrunner_test.go`:
```go
package taskrunner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"dfclientkit/account"
)

type testItem struct {
	Title string
}

func itemTitle(i testItem) string { return i.Title }

// fakeRequester 实现 Requester[testItem] 接口，用于测试。alwaysFail 为 true 时永远失败。
type fakeRequester struct {
	mu         sync.Mutex
	alwaysFail bool
	calls      map[string]int
}

func (f *fakeRequester) Publish(ctx context.Context, acc account.Account, item testItem) (string, error) {
	f.mu.Lock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[acc.CK]++
	f.mu.Unlock()
	if f.alwaysFail {
		return "", errors.New("boom")
	}
	return "ok:" + acc.CK, nil
}

func (f *fakeRequester) callCount(ck string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[ck]
}

func testTargets(cks ...string) []IndexedAccount {
	out := make([]IndexedAccount, len(cks))
	for i, ck := range cks {
		out[i] = IndexedAccount{Index: i, Account: account.Account{CK: ck}}
	}
	return out
}

func collectEvents(t *testing.T, r *Runner[testItem], cfg RunConfig, gate *PauseGate, pool []IndexedAccount, items []testItem) []Event {
	t.Helper()
	var mu sync.Mutex
	var events []Event
	err := r.Run(context.Background(), cfg, gate, pool, items, itemTitle, func(e Event) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Run err: %v", err)
	}
	return events
}

func TestRun_SingleAttemptAllSuccess(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1", "ck2")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	var success int
	for _, e := range events {
		if e.Kind == EventAttemptSuccess {
			success++
		}
	}
	if success != 2 {
		t.Errorf("成功次数 = %d, want 2", success)
	}
}

func TestRun_PerAccountLoopsMultipleTimes(t *testing.T) {
	items := []testItem{{Title: "a"}, {Title: "b"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 3, FailSwitchCount: 10, CycleRounds: 1}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	var titles []string
	for _, e := range events {
		if e.Kind == EventAttemptStart {
			titles = append(titles, e.ItemLabel)
		}
	}
	want := []string{"a", "b", "a"} // 循环复用内容
	if len(titles) != len(want) {
		t.Fatalf("尝试次数 = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Errorf("titles[%d] = %q, want %q", i, titles[i], want[i])
		}
	}
}

func TestRun_FailSwitchStopsAccountEarly(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 10, FailSwitchCount: 2, CycleRounds: 1}
	fr := &fakeRequester{alwaysFail: true}
	events := collectEvents(t, New[testItem](fr, nil), cfg, nil, targets, items)

	if got := fr.callCount("ck1"); got != 2 {
		t.Errorf("应只尝试 2 次就换号, 实际调用 %d 次", got)
	}
	var switched int
	for _, e := range events {
		if e.Kind == EventAccountSwitch {
			switched++
		}
	}
	if switched != 1 {
		t.Errorf("应有一次换号事件, got %d", switched)
	}
}

func TestRun_FailThenSuccessResetsCounter(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 4, FailSwitchCount: 2, CycleRounds: 1}

	calls := 0
	var mu sync.Mutex
	requester := requesterFunc(func(ctx context.Context, acc account.Account, item testItem) (string, error) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 2 { // 第 2 次失败，其余成功；失败不连续所以不会触发换号
			return "", errors.New("boom")
		}
		return "ok", nil
	})
	events := collectEvents(t, New[testItem](requester, nil), cfg, nil, targets, items)

	var switched int
	for _, e := range events {
		if e.Kind == EventAccountSwitch {
			switched++
		}
	}
	if switched != 0 {
		t.Errorf("非连续失败不应触发换号, got %d 次换号", switched)
	}
	if calls != 4 {
		t.Errorf("应完整跑满 4 次, got %d", calls)
	}
}

func TestRun_MultipleRounds(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 3, RoundIntervalSec: 0}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	var starts, dones []int
	for _, e := range events {
		if e.Kind == EventRoundStart {
			starts = append(starts, e.Round)
		}
		if e.Kind == EventRoundDone {
			dones = append(dones, e.Round)
		}
	}
	if len(starts) != 3 || len(dones) != 3 {
		t.Fatalf("应有 3 轮 start/done, got starts=%v dones=%v", starts, dones)
	}
	for i, want := range []int{1, 2, 3} {
		if starts[i] != want || dones[i] != want {
			t.Errorf("第 %d 轮编号不对: starts=%v dones=%v", i, starts, dones)
		}
	}
}

func TestRun_ConcurrencyProcessesAllAccounts(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1", "ck2", "ck3", "ck4", "ck5", "ck6")
	cfg := RunConfig{Threads: 4, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	seen := map[int]bool{}
	for _, e := range events {
		if e.Kind == EventAttemptSuccess {
			seen[e.AccountIndex] = true
		}
	}
	if len(seen) != 6 {
		t.Errorf("应处理全部 6 个账号, got %d: %v", len(seen), seen)
	}
}

func TestRun_CreateRepoFailureSwitchesWithoutAttempt(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 5, FailSwitchCount: 5, CycleRounds: 1, CreateRepo: true}
	fr := &fakeRequester{}
	events := collectEvents(t, New[testItem](fr, repoCreatorFunc(func(ctx context.Context, acc account.Account) error {
		return errors.New("no space")
	})), cfg, nil, targets, items)

	if fr.callCount("ck1") != 0 {
		t.Errorf("建仓库失败不应尝试处理, 实际调用 %d 次", fr.callCount("ck1"))
	}
	var switched, attempts int
	for _, e := range events {
		if e.Kind == EventAccountSwitch {
			switched++
		}
		if e.Kind == EventAttemptStart || e.Kind == EventAttemptSuccess || e.Kind == EventAttemptFailure {
			attempts++
		}
	}
	if switched != 1 {
		t.Errorf("应只有一次换号事件, got %d: %+v", switched, events)
	}
	if attempts != 0 {
		t.Errorf("建仓库失败不应产生任何处理尝试事件, got %d", attempts)
	}
}

func TestRun_PauseBlocksUntilResumed(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	gate := NewPauseGate()
	gate.Pause()

	events := make(chan Event, 8)
	done := make(chan error, 1)
	go func() {
		done <- New[testItem](&fakeRequester{}, nil).Run(context.Background(), cfg, gate, targets, items, itemTitle, func(e Event) {
			events <- e
		})
	}()

	// 轮次开始事件不受暂停影响（在派发 worker 之前就会发出），但暂停期间不应有任何处理尝试事件。
	select {
	case e := <-events:
		if e.Kind != EventRoundStart {
			t.Fatalf("暂停期间只应看到 EventRoundStart, got %+v", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("迟迟没有收到 EventRoundStart")
	}
	select {
	case e := <-events:
		t.Fatalf("暂停期间不应产生处理尝试事件, got %+v", e)
	case <-time.After(150 * time.Millisecond):
	}

	gate.Resume()

	select {
	case e := <-events:
		if e.Kind != EventAttemptStart {
			t.Errorf("恢复后第一个事件应为 EventAttemptStart, got %v", e.Kind)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("恢复后长时间没有事件")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run 迟迟未结束")
	}
}

func TestRun_CancelDuringWorkStopsProcessing(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1", "ck2", "ck3")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	ctx, cancel := context.WithCancel(context.Background())

	var mu sync.Mutex
	var started int
	err := New[testItem](&fakeRequester{}, nil).Run(ctx, cfg, nil, targets, items, itemTitle, func(e Event) {
		if e.Kind == EventAttemptStart {
			mu.Lock()
			started++
			mu.Unlock()
			cancel()
		}
	})
	if err == nil {
		t.Errorf("取消后 Run 应返回 context 错误")
	}
	if started != 1 {
		t.Errorf("取消后不应再开始新账号, started=%d", started)
	}
}

func TestRun_NoContentOrAccountsIsNoop(t *testing.T) {
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	r := New[testItem](&fakeRequester{}, nil)

	if err := r.Run(context.Background(), cfg, nil, testTargets("ck1"), nil, itemTitle, func(Event) {
		t.Fatal("没有内容时不应产生事件")
	}); err != nil {
		t.Fatalf("Run err: %v", err)
	}
	if err := r.Run(context.Background(), cfg, nil, nil, []testItem{{Title: "a"}}, itemTitle, func(Event) {
		t.Fatal("没有账号时不应产生事件")
	}); err != nil {
		t.Fatalf("Run err: %v", err)
	}
}

type requesterFunc func(ctx context.Context, acc account.Account, item testItem) (string, error)

func (f requesterFunc) Publish(ctx context.Context, acc account.Account, item testItem) (string, error) {
	return f(ctx, acc, item)
}

type repoCreatorFunc func(ctx context.Context, acc account.Account) error

func (f repoCreatorFunc) CreateSpace(ctx context.Context, acc account.Account) error {
	return f(ctx, acc)
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./taskrunner/...`
Expected: FAIL，编译错误（`undefined: RunConfig` / `undefined: New` / `undefined: IndexedAccount` 等，因为 `taskrunner.go` 还不存在）。

- [ ] **Step 3: 写实现**

`~/GolandProjects/dfclientkit/go/taskrunner/taskrunner.go`:
```go
// Package taskrunner 提供"用账号队列并发处理内容"的通用任务引擎：多线程并发处理
// 账号池，单个账号可循环处理多次，连续失败达到阈值换号，整个账号池可循环多轮，
// 支持暂停/恢复与取消，通过回调把进度事件传给上层。具体处理协议（Requester）与
// 处理对象类型（Item，比如文章、视频、评论文本）由各消费方注入，本包不关心。
package taskrunner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"dfclientkit/account"
)

// Requester 执行一次"用某账号处理某个 Item"的请求，返回结果描述或错误。
type Requester[Item any] interface {
	Publish(ctx context.Context, acc account.Account, item Item) (string, error)
}

// RepoCreator 在处理某个账号前视需要建一个"仓库/空间"。
type RepoCreator interface {
	CreateSpace(ctx context.Context, acc account.Account) error
}

// EventKind 标识一次处理过程中的事件类型。
type EventKind int

const (
	EventAttemptStart   EventKind = iota // 某账号的一次处理尝试开始
	EventAttemptSuccess                  // 该次尝试成功
	EventAttemptFailure                  // 该次尝试失败
	EventAccountSwitch                   // 放弃当前账号，换下一个（达到每号处理上限/连续失败换号/建仓库失败）
	EventRoundStart                      // 新的一轮开始
	EventRoundProgress                   // 本轮进度更新
	EventRoundDone                       // 本轮结束
)

// Event 是回传给上层的进度事件。
type Event struct {
	Kind         EventKind
	AccountIndex int    // 账号在原始队列中的下标；EventRoundStart/RoundProgress/RoundDone 不适用
	CK           string
	ItemLabel    string // 本次处理对象的展示文案（比如文章标题），由调用方通过 itemLabel 提供
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

// RunConfig 是一次批量任务的运行参数。
type RunConfig struct {
	Threads          int  // 并发线程数
	IntervalSec      int  // 同一账号相邻两次处理尝试之间的等待秒数
	PerAccountCount  int  // 单个账号最多处理多少次
	FailSwitchCount  int  // 账号连续失败达到此次数就换号
	CycleRounds      int  // 账号池整体循环轮数
	RoundIntervalSec int  // 相邻两轮之间的等待秒数
	CreateRepo       bool // 处理账号前是否先建仓库/空间
}

// Normalize 纠正非法数值，供消费方在保存/发起任务前调用，也在 Run 内部兜底调用。
func (c *RunConfig) Normalize() {
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

// Runner 并发执行账号处理任务，Item 是被处理对象的类型（文章、视频……由消费方决定）。
type Runner[Item any] struct {
	client Requester[Item]
	repo   RepoCreator
}

// New 创建 Runner；repo 为 nil 时忽略"创建仓库"选项。
func New[Item any](client Requester[Item], repo RepoCreator) *Runner[Item] {
	return &Runner[Item]{client: client, repo: repo}
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

// Run 执行账号池的批量处理任务。gate 为 nil 时视为从不暂停。items 为空时直接返回。
// itemLabel 从一个 Item 里提取用于事件展示的文案（比如文章标题）。
func (r *Runner[Item]) Run(ctx context.Context, cfg RunConfig, gate *PauseGate, pool []IndexedAccount, items []Item, itemLabel func(Item) string, onEvent func(Event)) error {
	if len(items) == 0 || len(pool) == 0 {
		return nil
	}
	cfg.Normalize()
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
					r.runAccount(ctx, cfg, gate, ia, items, itemLabel, onEvent)
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

func (r *Runner[Item]) runAccount(ctx context.Context, cfg RunConfig, gate *PauseGate, ia IndexedAccount, items []Item, itemLabel func(Item) string, onEvent func(Event)) {
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

		item := items[i%len(items)]
		label := itemLabel(item)
		onEvent(Event{Kind: EventAttemptStart, AccountIndex: ia.Index, CK: ia.Account.CK, ItemLabel: label})

		result, err := r.client.Publish(ctx, ia.Account, item)
		if err != nil {
			consecFail++
			onEvent(Event{Kind: EventAttemptFailure, AccountIndex: ia.Index, CK: ia.Account.CK, ItemLabel: label, Err: err})
			if consecFail >= cfg.FailSwitchCount {
				onEvent(Event{Kind: EventAccountSwitch, AccountIndex: ia.Index, CK: ia.Account.CK, Err: errors.New("连续失败达到换号阈值")})
				return
			}
		} else {
			consecFail = 0
			onEvent(Event{Kind: EventAttemptSuccess, AccountIndex: ia.Index, CK: ia.Account.CK, ItemLabel: label, Result: result})
		}

		if i < cfg.PerAccountCount-1 && cfg.IntervalSec > 0 {
			if err := sleepCtx(ctx, time.Duration(cfg.IntervalSec)*time.Second); err != nil {
				return
			}
		}
	}
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./taskrunner/... -v`
Expected: PASS，11 个测试全绿（并发/暂停相关的测试用了真实的 `time.Sleep`/`time.After`，跑起来会花几秒钟，属于正常现象）。

- [ ] **Step 5: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add go/taskrunner
git commit -m "feat: taskrunner 包——账号队列并发任务引擎"
```

---

### Task 5: runlog 包

**Files:**
- Create: `~/GolandProjects/dfclientkit/go/runlog/runlog.go`
- Test: `~/GolandProjects/dfclientkit/go/runlog/runlog_test.go`

- [ ] **Step 1: 写测试文件**

`~/GolandProjects/dfclientkit/go/runlog/runlog_test.go`:
```go
package runlog

import (
	"errors"
	"testing"

	"dfclientkit/taskrunner"
)

func TestTagForKindsAndHighlight(t *testing.T) {
	cases := []struct {
		kind     taskrunner.EventKind
		wantTag  string
		wantKind Kind
		wantHi   bool
	}{
		{taskrunner.EventAttemptStart, "[开始]", KindStart, false},
		{taskrunner.EventAttemptSuccess, "[成功]", KindSuccess, false},
		{taskrunner.EventAttemptFailure, "[失败]", KindFailure, true},
		{taskrunner.EventAccountSwitch, "[换号]", KindSwitch, true},
		{taskrunner.EventRoundStart, "[轮次]", KindInfo, false},
		{taskrunner.EventRoundDone, "[轮次]", KindInfo, false},
	}
	for _, c := range cases {
		tag, kind, hi := TagFor(c.kind)
		if tag != c.wantTag || kind != c.wantKind || hi != c.wantHi {
			t.Errorf("TagFor(%v) = %q/%v/%v, want %q/%v/%v", c.kind, tag, kind, hi, c.wantTag, c.wantKind, c.wantHi)
		}
	}
}

func TestLineForFormattingWithVerb(t *testing.T) {
	cases := []struct {
		event taskrunner.Event
		verb  string
		want  string
	}{
		{taskrunner.Event{Kind: taskrunner.EventAttemptStart, CK: "ck1", ItemLabel: "hello"}, "发布", "账号 ck1 开始发布《hello》"},
		{taskrunner.Event{Kind: taskrunner.EventAttemptSuccess, CK: "ck1", ItemLabel: "hello", Result: "ok"}, "发布", "账号 ck1 发布《hello》成功: ok"},
		{taskrunner.Event{Kind: taskrunner.EventAttemptFailure, CK: "ck1", ItemLabel: "hello", Err: errors.New("boom")}, "发布", "账号 ck1 发布《hello》失败: boom"},
		{taskrunner.Event{Kind: taskrunner.EventAccountSwitch, CK: "ck1", Err: errors.New("连续失败达到换号阈值")}, "发布", "账号 ck1 换号: 连续失败达到换号阈值"},
		{taskrunner.Event{Kind: taskrunner.EventRoundStart, Round: 2, RoundTotal: 5}, "发布", "第 2 轮开始，共 5 个账号"},
		{taskrunner.Event{Kind: taskrunner.EventRoundDone, Round: 2}, "发布", "第 2 轮结束"},
	}
	for _, c := range cases {
		if got := LineFor(c.event, c.verb); got != c.want {
			t.Errorf("LineFor(%+v, %q) = %q, want %q", c.event, c.verb, got, c.want)
		}
	}
}

func TestLineForDefaultsVerbToProcess(t *testing.T) {
	e := taskrunner.Event{Kind: taskrunner.EventAttemptStart, CK: "ck1", ItemLabel: "hello"}
	want := "账号 ck1 开始处理《hello》"
	if got := LineFor(e, ""); got != want {
		t.Errorf("LineFor(_, \"\") = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./runlog/...`
Expected: FAIL，编译错误 `undefined: TagFor` / `undefined: LineFor` / `undefined: Kind`。

- [ ] **Step 3: 写实现**

`~/GolandProjects/dfclientkit/go/runlog/runlog.go`:
```go
// Package runlog 把 taskrunner 的处理事件格式化为运行日志文案，供前端渲染着色。
package runlog

import (
	"fmt"

	"dfclientkit/taskrunner"
)

// Kind 标识一条日志在前端应使用的着色分类，前端据此映射到主题色。
type Kind string

const (
	KindStart   Kind = "start"
	KindSuccess Kind = "success"
	KindFailure Kind = "failure"
	KindSwitch  Kind = "switch"
	KindInfo    Kind = "info"
)

// TagFor 返回某类事件的日志标签、着色分类，以及正文是否也要跟着上色
// （失败/换号的正文本身也标红/标黄，其余类型正文用默认前景色）。
// EventRoundProgress 不产生日志行，调用方应在此之前过滤掉。
func TagFor(k taskrunner.EventKind) (tag string, kind Kind, highlightMessage bool) {
	switch k {
	case taskrunner.EventAttemptStart:
		return "[开始]", KindStart, false
	case taskrunner.EventAttemptSuccess:
		return "[成功]", KindSuccess, false
	case taskrunner.EventAttemptFailure:
		return "[失败]", KindFailure, true
	case taskrunner.EventAccountSwitch:
		return "[换号]", KindSwitch, true
	case taskrunner.EventRoundStart, taskrunner.EventRoundDone:
		return "[轮次]", KindInfo, false
	default:
		return "[信息]", KindInfo, false
	}
}

// LineFor 把一条事件格式化为日志正文（不含时间戳/标签）。verb 是这次任务的业务
// 动作动词（比如"发布"/"注册"/"采集"），传空字符串时用缺省的"处理"。
func LineFor(e taskrunner.Event, verb string) string {
	if verb == "" {
		verb = "处理"
	}
	switch e.Kind {
	case taskrunner.EventAttemptStart:
		return fmt.Sprintf("账号 %s 开始%s《%s》", e.CK, verb, e.ItemLabel)
	case taskrunner.EventAttemptSuccess:
		return fmt.Sprintf("账号 %s %s《%s》成功: %s", e.CK, verb, e.ItemLabel, e.Result)
	case taskrunner.EventAttemptFailure:
		return fmt.Sprintf("账号 %s %s《%s》失败: %v", e.CK, verb, e.ItemLabel, e.Err)
	case taskrunner.EventAccountSwitch:
		return fmt.Sprintf("账号 %s 换号: %v", e.CK, e.Err)
	case taskrunner.EventRoundStart:
		return fmt.Sprintf("第 %d 轮开始，共 %d 个账号", e.Round, e.RoundTotal)
	case taskrunner.EventRoundDone:
		return fmt.Sprintf("第 %d 轮结束", e.Round)
	default:
		return e.CK
	}
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd ~/GolandProjects/dfclientkit/go && go test ./runlog/... -v`
Expected: PASS，3 个测试全绿。

- [ ] **Step 5: 跑全部 dfclientkit Go 测试确认整体健康**

Run: `cd ~/GolandProjects/dfclientkit/go && go vet ./... && go test ./...`
Expected: `go vet` 无输出；四个包（account/appconfig/taskrunner/runlog）全部 `ok`。

- [ ] **Step 6: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add go/runlog
git commit -m "feat: runlog 包——事件转日志文案，动词可配置"
```

---

## Phase B — dfclientkit/ui-shell 前端基建

### Task 6: ui-shell 包骨架 + 基础主题

**Files:**
- Create: `~/GolandProjects/dfclientkit/ui-shell/package.json`
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/index.js`
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/theme.css`

- [ ] **Step 1: 写 package.json**

`~/GolandProjects/dfclientkit/ui-shell/package.json`:
```json
{
  "name": "@dongfang/df-ui-shell",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "main": "./src/index.js",
  "exports": {
    ".": "./src/index.js",
    "./theme.css": "./src/theme.css"
  },
  "peerDependencies": {
    "vue": "^3.4.0"
  }
}
```

- [ ] **Step 2: 写 index.js（统一导出六个组件）**

`~/GolandProjects/dfclientkit/ui-shell/src/index.js`:
```js
export { default as TitleBar } from './TitleBar.vue'
export { default as NavRail } from './NavRail.vue'
export { default as StatusBar } from './StatusBar.vue'
export { default as LogPanel } from './LogPanel.vue'
export { default as NumberStepper } from './NumberStepper.vue'
export { default as ResultsModal } from './ResultsModal.vue'
```

- [ ] **Step 3: 写基础主题（从 githubbaidu 现有 theme.css 原样搬迁，只去掉一处引用具体历史设计稿的注释）**

`~/GolandProjects/dfclientkit/ui-shell/src/theme.css`:
```css
/* 设计令牌：桌面壳基础主题的 :root / [data-theme="dark"] 配色变量。 */
:root {
  --bg: #eef0f3;
  --surface: #ffffff;
  --surface-2: #f6f7f9;
  --nav: #f3f5f7;
  --border: #dce0e6;
  --border-strong: #b6bdc7;
  --text: #1a1e23;
  --muted: #6e7681;
  --accent: #1f6feb;
  --accent-weak: #e7f0fe;
  --ok: #1a7f37;
  --warn: #9a6700;
  --err: #cf222e;
  --log-bg: #0b0e12;
  --shadow: 0 18px 48px rgba(16, 22, 32, 0.18);
}

[data-theme="dark"] {
  --bg: #0a0d12;
  --surface: #161b22;
  --surface-2: #1b2129;
  --nav: #11161d;
  --border: #2a313a;
  --border-strong: #3d444d;
  --text: #e6edf3;
  --muted: #8b949e;
  --accent: #388bfd;
  --accent-weak: #132a4d;
  --ok: #3fb950;
  --warn: #d29922;
  --err: #f85149;
  --log-bg: #05070a;
  --shadow: 0 18px 48px rgba(0, 0, 0, 0.55);
}

* {
  box-sizing: border-box;
}

html,
body,
#app {
  height: 100%;
}

body {
  margin: 0;
  background: var(--bg);
  color: var(--text);
  font-family: system-ui, "PingFang SC", "Microsoft YaHei", "Helvetica Neue", sans-serif;
  -webkit-font-smoothing: antialiased;
  overflow: hidden;
}

.mono {
  font-family: ui-monospace, "SF Mono", "Cascadia Mono", Consolas, monospace;
}

::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

::-webkit-scrollbar-thumb {
  background: var(--border-strong);
  border-radius: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.25;
  }
}
```

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add ui-shell/package.json ui-shell/src/index.js ui-shell/src/theme.css
git commit -m "feat: ui-shell 包骨架 + 基础主题"
```

---

### Task 7: StatusBar / LogPanel / NumberStepper（原样搬迁）

这三个组件在 githubbaidu 里已经是纯 props/emit 驱动、没有业务耦合，原样复制过来，不改一行逻辑。

**Files:**
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/StatusBar.vue`
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/LogPanel.vue`
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/NumberStepper.vue`

- [ ] **Step 1: 复制 StatusBar.vue**

Run:
```bash
cp ~/GolandProjects/githubbaidu/frontend/src/components/StatusBar.vue \
   ~/GolandProjects/dfclientkit/ui-shell/src/StatusBar.vue
```

- [ ] **Step 2: 复制 LogPanel.vue**

Run:
```bash
cp ~/GolandProjects/githubbaidu/frontend/src/components/LogPanel.vue \
   ~/GolandProjects/dfclientkit/ui-shell/src/LogPanel.vue
```

- [ ] **Step 3: 复制 NumberStepper.vue**

Run:
```bash
cp ~/GolandProjects/githubbaidu/frontend/src/components/NumberStepper.vue \
   ~/GolandProjects/dfclientkit/ui-shell/src/NumberStepper.vue
```

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add ui-shell/src/StatusBar.vue ui-shell/src/LogPanel.vue ui-shell/src/NumberStepper.vue
git commit -m "feat: ui-shell 原样搬迁 StatusBar/LogPanel/NumberStepper"
```

---

### Task 8: TitleBar 组件改造

原组件直接 `import { WindowMinimise, WindowToggleMaximise, Quit } from '../../wailsjs/runtime/runtime'`——这是githubbaidu 项目里 Wails 生成的本地绑定文件，路径是这个项目专属的，搬到 ui-shell 后无法解析（每个消费方的 `wailsjs` 生成目录位置都不一样）。改成三个窗口控制按钮改为 `emit`，由各消费方自己接住并调用自己项目里的 wails 绑定；应用名 `GitHub 文章发布器` 和 logo 字母 `G` 改成 `appName`/`logoText` prop。

**Files:**
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/TitleBar.vue`

- [ ] **Step 1: 写组件**

`~/GolandProjects/dfclientkit/ui-shell/src/TitleBar.vue`:
```vue
<script setup>
defineProps({
  appName: { type: String, required: true },
  logoText: { type: String, required: true },
  theme: { type: String, required: true }, // 'light' | 'dark'
  version: { type: String, required: true },
  pillText: { type: String, required: true },
  pillKind: { type: String, required: true }, // 'muted' | 'running' | 'success' | 'error'
})
const emit = defineEmits(['toggle-theme', 'minimize', 'maximize', 'close'])
</script>

<template>
  <div class="titlebar">
    <div class="logo">{{ logoText }}</div>
    <span class="app-name">{{ appName }}</span>
    <span class="version-pill">{{ version }}</span>
    <div class="spacer" />
    <div class="pill" :class="`pill-${pillKind}`">
      <span class="pill-dot" />
      <span class="pill-text">{{ pillText }}</span>
    </div>
    <button class="icon-btn theme-toggle" :title="theme === 'light' ? '切换为深色' : '切换为浅色'" @click="emit('toggle-theme')">
      <svg v-if="theme === 'light'" width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <path d="M14 9.2A6 6 0 016.8 2 6.5 6.5 0 1014 9.2z" />
      </svg>
      <svg v-else width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
        <circle cx="8" cy="8" r="3.2" /><path d="M8 1v1.6M8 13.4V15M1 8h1.6M13.4 8H15M3 3l1.1 1.1M11.9 11.9L13 13M13 3l-1.1 1.1M4.1 11.9L3 13" />
      </svg>
    </button>
    <div class="window-controls">
      <button class="icon-btn" title="最小化" @click="emit('minimize')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <path d="M3 12h10" stroke-linecap="round" />
        </svg>
      </button>
      <button class="icon-btn" title="最大化 / 还原" @click="emit('maximize')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <rect x="3.5" y="3.5" width="9" height="9" rx="1.5" />
        </svg>
      </button>
      <button class="icon-btn close-btn" title="关闭" @click="emit('close')">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
          <path d="M4 4l8 8M12 4l-8 8" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.titlebar {
  --wails-draggable: drag;
  height: 52px;
  flex: 0 0 52px;
  margin: 10px 10px 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 10px 0 14px;
  background: var(--nav);
  border: 1px solid var(--border);
  border-radius: 14px;
  box-shadow: 0 4px 16px rgba(16, 22, 32, 0.1);
}
.logo {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  font-family: ui-monospace, "SF Mono", Consolas, monospace;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 24px;
}
.app-name {
  font-size: 14.5px;
  font-weight: 700;
}
.version-pill {
  padding: 3px 8px;
  border-radius: 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--muted);
  font-size: 11px;
  font-weight: 500;
  font-family: ui-monospace, "SF Mono", Consolas, monospace;
}
.spacer {
  flex: 1;
}
.pill {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 5px 11px;
  border-radius: 14px;
  font-size: 12px;
  font-weight: 500;
}
.pill-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
.pill-muted {
  background: var(--surface);
  color: var(--muted);
}
.pill-running {
  background: rgba(63, 185, 80, 0.14);
  color: var(--ok);
}
.pill-running .pill-dot {
  animation: pulse 1.4s ease-in-out infinite;
}
.pill-success {
  background: rgba(31, 111, 235, 0.12);
  color: var(--accent);
}
.pill-error {
  background: rgba(248, 81, 73, 0.14);
  color: var(--err);
}
.icon-btn {
  --wails-draggable: no-drag;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  flex: 0 0 28px;
}
.icon-btn:hover {
  color: var(--text);
  border-color: var(--border-strong);
  background: var(--surface);
}
.window-controls {
  display: flex;
  align-items: center;
  gap: 6px;
}
.close-btn:hover {
  color: #fff;
  border-color: var(--err);
  background: var(--err);
}
</style>
```

- [ ] **Step 2: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add ui-shell/src/TitleBar.vue
git commit -m "feat: ui-shell TitleBar 组件——appName/logoText 可配置，窗口控制改为 emit"
```

---

### Task 9: NavRail 组件改造

原组件写死了 7 个固定菜单项和对应 SVG。改成 `items`/`bottomItems`（原来底部单独的"使用说明"按钮也纳入 `bottomItems`）从外部传入，`icon` 是可选的 Vue 组件，用 `<component :is>` 渲染。

**Files:**
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/NavRail.vue`

- [ ] **Step 1: 写组件**

`~/GolandProjects/dfclientkit/ui-shell/src/NavRail.vue`:
```vue
<script setup>
defineProps({
  page: { type: String, required: true },
  items: { type: Array, required: true }, // [{ key, cn, en, icon? }]
  bottomItems: { type: Array, default: () => [] }, // 渲染在下方（比如"使用说明"）
})
const emit = defineEmits(['navigate'])
</script>

<template>
  <div class="nav">
    <div class="nav-title">工作台 WORKSPACE</div>

    <button v-for="item in items" :key="item.key" class="nav-item" :class="{ active: page === item.key }" @click="emit('navigate', item.key)">
      <component :is="item.icon" v-if="item.icon" />
      <span class="nav-label">
        <span class="nav-label-cn">{{ item.cn }}</span>
        <span class="nav-label-en">{{ item.en }}</span>
      </span>
    </button>

    <div class="nav-spacer" />

    <button v-for="item in bottomItems" :key="item.key" class="nav-item" :class="{ active: page === item.key }" @click="emit('navigate', item.key)">
      <component :is="item.icon" v-if="item.icon" />
      <span class="nav-label">
        <span class="nav-label-cn">{{ item.cn }}</span>
        <span class="nav-label-en">{{ item.en }}</span>
      </span>
    </button>
  </div>
</template>

<style scoped>
.nav {
  width: 182px;
  flex: 0 0 182px;
  background: var(--nav);
  border-right: 1px solid var(--border);
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}
.nav-title {
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--muted);
  padding: 2px 8px 6px;
  text-transform: uppercase;
}
.nav-spacer {
  flex: 1;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  flex: 0 0 40px;
  padding: 0 10px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  border-left: 3px solid transparent;
}
.nav-item svg {
  flex: 0 0 15px;
}
.nav-item:hover {
  background: var(--surface);
}
.nav-item.active {
  background: var(--accent-weak);
  color: var(--accent);
  border-left-color: var(--accent);
}
.nav-label {
  display: flex;
  flex-direction: column;
  line-height: 1.1;
  overflow: hidden;
}
.nav-label-cn {
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}
.nav-label-en {
  font-size: 9px;
  letter-spacing: 0.05em;
  color: var(--muted);
  white-space: nowrap;
}
.nav-item.active .nav-label-en {
  color: var(--accent);
  opacity: 0.75;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add ui-shell/src/NavRail.vue
git commit -m "feat: ui-shell NavRail 组件——菜单项与图标改为外部传入"
```

---

### Task 10: ResultsModal 组件改造

结果列表原来写死"CK/标题/结果"三列的字段与样式类，标题写死"发布结果"。改成 `columns`（含每列的 `class`，保证默认值渲染出来的 HTML 和迁移前逐字节一致）和 `title` 两个可配置 prop，默认值维持现状，githubbaidu 现有调用代码不用改。

**Files:**
- Create: `~/GolandProjects/dfclientkit/ui-shell/src/ResultsModal.vue`

- [ ] **Step 1: 写组件**

`~/GolandProjects/dfclientkit/ui-shell/src/ResultsModal.vue`:
```vue
<script setup>
defineProps({
  results: { type: Array, required: true }, // 每行对象的字段由 columns 决定
  title: { type: String, default: '发布结果' },
  columns: {
    type: Array,
    default: () => [
      { key: 'ck', label: 'CK', class: 'mono' },
      { key: 'title', label: '标题', class: 'ellipsis' },
      { key: 'value', label: '结果', class: 'ellipsis mono' },
    ],
  },
})
const emit = defineEmits(['close', 'copy-all'])
</script>

<template>
  <div class="overlay" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">{{ title }}</span>
        <span class="modal-count mono">{{ results.length }} 条</span>
        <div class="spacer" />
        <button class="btn-outline" @click="emit('copy-all')">复制全部</button>
        <button class="btn-close" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div v-if="results.length === 0" class="empty">暂无成功结果</div>
        <div v-for="(r, i) in results" :key="i" class="result-row">
          <span class="mono muted">{{ r.time }}</span>
          <span v-for="col in columns" :key="col.key" :class="col.class">{{ r[col.key] }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}
.modal {
  width: 560px;
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  box-shadow: var(--shadow);
  overflow: hidden;
}
.modal-header {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--border);
}
.modal-title {
  font-size: 14px;
  font-weight: 700;
}
.modal-count {
  color: var(--muted);
  font-size: 12px;
}
.spacer {
  flex: 1;
}
.btn-outline {
  height: 28px;
  padding: 0 10px;
  border-radius: 5px;
  border: 1px solid var(--border-strong);
  background: var(--surface);
  font-size: 12px;
  cursor: pointer;
  color: var(--text);
}
.btn-outline:hover {
  border-color: var(--accent);
  color: var(--accent);
}
.btn-close {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 16px;
  cursor: pointer;
  border-radius: 5px;
}
.btn-close:hover {
  color: var(--text);
  background: var(--surface-2);
}
.modal-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 6px 14px;
}
.result-row {
  display: grid;
  grid-template-columns: 76px 140px 1fr 1fr;
  gap: 10px;
  padding: 7px 0;
  border-bottom: 1px solid var(--border);
  font-size: 12.5px;
  align-items: center;
}
.result-row:last-child {
  border-bottom: none;
}
.muted {
  color: var(--muted);
}
.ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.empty {
  padding: 30px 0;
  text-align: center;
  color: var(--muted);
  font-size: 12.5px;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
cd ~/GolandProjects/dfclientkit
git add ui-shell/src/ResultsModal.vue
git commit -m "feat: ui-shell ResultsModal 组件——列与标题改为可配置"
```

---

## Phase C — githubbaidu 接入 dfclientkit（Go 侧）

### Task 11: go.mod 接入 dfclientkit

**Files:**
- Modify: `githubbaidu/go.mod`

- [ ] **Step 1: 加 require + replace**

在 `githubbaidu/go.mod` 的 `require github.com/wailsapp/wails/v2 v2.15.0` 那一行下面新增一段：
```
require dfclientkit v0.0.0

replace dfclientkit => ../dfclientkit/go
```

完整效果（文件顶部几行）：
```
module githubbaidu

go 1.26

require github.com/wailsapp/wails/v2 v2.15.0

require dfclientkit v0.0.0

replace dfclientkit => ../dfclientkit/go

require (
	git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3 // indirect
	...
)
```

- [ ] **Step 2: 跑 go mod tidy 确认可解析**

Run: `cd ~/GolandProjects/githubbaidu && go mod tidy`
Expected: 命令成功退出，`go.mod` 里 `dfclientkit` 那行保持不变（本地 replace 不需要写具体 hash/版本校验，也不会生成 `go.sum` 条目）。

- [ ] **Step 3: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add go.mod go.sum
git commit -m "chore: 接入 dfclientkit（本地 replace）"
```

---

### Task 12: internal/account 改薄壳

原来的实现整体搬到了 `dfclientkit/account`，这里只留一层委托 + 本项目专属的 `DefaultPath()`（`dfclientkit/account` 不知道"ghpublisher"这个 app 名）。原来的单测覆盖的逻辑现在在 dfclientkit 里测过了，这里只保留 `DefaultPath()` 这一处本项目专属逻辑的测试。

**Files:**
- Modify: `githubbaidu/internal/account/account.go`
- Modify: `githubbaidu/internal/account/account_test.go`

- [ ] **Step 1: 重写 account.go**

`githubbaidu/internal/account/account.go`（完整替换原内容）:
```go
// Package account 是 dfclientkit/account 在本项目内的别名层，保留
// githubbaidu/internal/account 这个导入路径不变，方便调用方少改代码；同时提供
// 本项目专属的默认存盘路径。
package account

import (
	"path/filepath"

	dfaccount "dfclientkit/account"
	"dfclientkit/appconfig"
)

type Account = dfaccount.Account

const (
	StatusPending = dfaccount.StatusPending
	StatusRunning = dfaccount.StatusRunning
	StatusSuccess = dfaccount.StatusSuccess
	StatusFailed  = dfaccount.StatusFailed
)

var ParseImportText = dfaccount.ParseImportText

const appName = "ghpublisher"

// DefaultPath 返回账号队列在当前系统上的默认存放路径。
func DefaultPath() (string, error) {
	dir, err := appconfig.DefaultDir(appName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "accounts.json"), nil
}

func Load(path string) []Account             { return dfaccount.Load(path) }
func Save(path string, list []Account) error { return dfaccount.Save(path, list) }
```

- [ ] **Step 2: 重写 account_test.go（只测本项目专属的 DefaultPath）**

`githubbaidu/internal/account/account_test.go`（完整替换原内容）:
```go
package account

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPathEndsWithAccountsJSON(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if filepath.Base(path) != "accounts.json" {
		t.Errorf("DefaultPath() = %q, want basename accounts.json", path)
	}
	if !strings.Contains(path, appName) {
		t.Errorf("DefaultPath() = %q, want it to contain app dir %q", path, appName)
	}
}
```

- [ ] **Step 3: 跑测试确认通过**

Run: `cd ~/GolandProjects/githubbaidu && go test ./internal/account/... -v`
Expected: PASS。

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add internal/account
git commit -m "refactor: internal/account 改为 dfclientkit 薄壳"
```

---

### Task 13: internal/config 改薄壳

`Config` 内嵌 `taskrunner.RunConfig`（拿到 `Threads`/`IntervalSec`/`PerAccountCount`/`FailSwitchCount`/`CycleRounds`/`RoundIntervalSec`/`CreateRepo` 这些通用字段，通过 Go 的字段提升，`cfg.Threads` 这种访问方式不用改），本项目只加 `KeywordSlots`。`Load`/`Save` 委托给 `appconfig`，不再需要自己算路径——原来 `app.go` 里 `LoadConfig`/`SaveConfig` 手动拼 `DefaultPath()` 的那部分逻辑现在可以简化掉，放在 Task 16 一起改。

**Files:**
- Modify: `githubbaidu/internal/config/config.go`
- Modify: `githubbaidu/internal/config/config_test.go`

- [ ] **Step 1: 重写 config.go**

`githubbaidu/internal/config/config.go`（完整替换原内容）:
```go
// Package config 保存发布任务的运行参数，通用部分委托给 dfclientkit。
package config

import (
	"dfclientkit/appconfig"
	"dfclientkit/taskrunner"
)

const appName = "ghpublisher"

// Config 保存发布任务的运行参数。不依赖任何 UI 类型。
type Config struct {
	taskrunner.RunConfig
	KeywordSlots int `json:"keywordSlots"` // 内容里的关键词插入位数量
}

// Normalize 纠正非法数值。
func (c *Config) Normalize() {
	c.RunConfig.Normalize()
	if c.KeywordSlots < 0 {
		c.KeywordSlots = 0
	}
}

func defaults() Config {
	return Config{
		RunConfig: taskrunner.RunConfig{
			Threads:          1,
			IntervalSec:      1,
			PerAccountCount:  1,
			FailSwitchCount:  3,
			CycleRounds:      1,
			RoundIntervalSec: 1,
		},
	}
}

// Load 读取磁盘上保存的任务参数；不存在则返回默认值。
func Load() Config {
	return appconfig.Load(appName, defaults())
}

// Save 把任务参数写入磁盘。
func Save(cfg Config) error {
	return appconfig.Save(appName, cfg)
}
```

- [ ] **Step 2: 重写 config_test.go（只测本项目专属的 KeywordSlots 纠正逻辑；RunConfig 的纠正逻辑已经在 dfclientkit 测过）**

`githubbaidu/internal/config/config_test.go`（完整替换原内容）:
```go
package config

import "testing"

func TestNormalizeClampsKeywordSlots(t *testing.T) {
	cfg := Config{KeywordSlots: -5}
	cfg.Normalize()
	if cfg.KeywordSlots != 0 {
		t.Errorf("KeywordSlots = %d, want 0", cfg.KeywordSlots)
	}
}

func TestNormalizeKeepsPositiveKeywordSlots(t *testing.T) {
	cfg := Config{KeywordSlots: 3}
	cfg.Normalize()
	if cfg.KeywordSlots != 3 {
		t.Errorf("KeywordSlots = %d, want 3（合法值不应被改动）", cfg.KeywordSlots)
	}
}
```

- [ ] **Step 3: 跑测试确认通过**

Run: `cd ~/GolandProjects/githubbaidu && go test ./internal/config/... -v`
Expected: PASS。

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add internal/config
git commit -m "refactor: internal/config 改为 dfclientkit 薄壳，内嵌 RunConfig"
```

---

### Task 14: internal/accountpublish 改薄壳

用一个包一层的 `Runner` struct（不是类型别名——因为要在内部把 `itemLabel` 函数塞进去，类型别名做不到这一点）把 `taskrunner.Runner[article.Article]` 包起来，让 `Run(...)` 的调用签名和迁移前完全一样，`app.go` 那边的调用代码不用改。`TODORequester`/`TODORepoCreator` 这两个占位实现留在这里，不进基建。

**Files:**
- Modify: `githubbaidu/internal/accountpublish/accountpublish.go`
- Delete: `githubbaidu/internal/accountpublish/accountpublish_test.go`

- [ ] **Step 1: 重写 accountpublish.go**

`githubbaidu/internal/accountpublish/accountpublish.go`（完整替换原内容）:
```go
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
```

- [ ] **Step 2: 删除旧测试文件（并发/换号/多轮/暂停恢复的逻辑已经在 dfclientkit/taskrunner 测过；这层薄壳没有自己的独立逻辑）**

Run: `rm ~/GolandProjects/githubbaidu/internal/accountpublish/accountpublish_test.go`

- [ ] **Step 3: 编译确认通过（这层没有测试，靠编译器检查签名对不对）**

Run: `cd ~/GolandProjects/githubbaidu && go build ./internal/accountpublish/...`
Expected: 编译成功，无输出。

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add internal/accountpublish
git commit -m "refactor: internal/accountpublish 改为 dfclientkit/taskrunner 薄壳"
```

---

### Task 15: internal/runlog 改薄壳

**Files:**
- Modify: `githubbaidu/internal/runlog/runlog.go`
- Delete: `githubbaidu/internal/runlog/runlog_test.go`

- [ ] **Step 1: 重写 runlog.go**

`githubbaidu/internal/runlog/runlog.go`（完整替换原内容）:
```go
// Package runlog 把账号发布事件格式化为运行日志文案，供前端渲染着色。
// 通用的标签/文案格式委托给 dfclientkit/runlog，这里固定业务动词为"发布"。
package runlog

import (
	"dfclientkit/runlog"

	"githubbaidu/internal/accountpublish"
)

type Kind = runlog.Kind

const (
	KindStart   = runlog.KindStart
	KindSuccess = runlog.KindSuccess
	KindFailure = runlog.KindFailure
	KindSwitch  = runlog.KindSwitch
	KindInfo    = runlog.KindInfo
)

// TagFor 返回某类发布事件的日志标签、着色分类，以及正文是否也要跟着上色。
func TagFor(k accountpublish.EventKind) (tag string, kind Kind, highlightMessage bool) {
	return runlog.TagFor(k)
}

// LineFor 把一条发布事件格式化为日志正文（不含时间戳/标签）。
func LineFor(e accountpublish.Event) string {
	return runlog.LineFor(e, "发布")
}
```

- [ ] **Step 2: 删除旧测试文件（标签/文案格式化逻辑已经在 dfclientkit/runlog 测过；这层薄壳只是固定了一个 verb 字符串常量）**

Run: `rm ~/GolandProjects/githubbaidu/internal/runlog/runlog_test.go`

- [ ] **Step 3: 编译确认通过**

Run: `cd ~/GolandProjects/githubbaidu && go build ./internal/runlog/...`
Expected: 编译成功，无输出。

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add internal/runlog
git commit -m "refactor: internal/runlog 改为 dfclientkit/runlog 薄壳，固定动词为「发布」"
```

---

### Task 16: app.go 调整 + 全量 Go 侧验证

`LoadConfig`/`SaveConfig` 里原来手动拼 `config.DefaultPath()` 的逻辑现在可以去掉了（`config.Load()`/`config.Save(cfg)` 内部已经通过 `appconfig` 处理路径）。其余方法（账号相关、发布相关）因为薄壳层保持了原有签名，不用改。

**Files:**
- Modify: `githubbaidu/app.go:92-113`（`LoadConfig`/`SaveConfig` 两个方法）

- [ ] **Step 1: 简化 LoadConfig/SaveConfig**

把 `app.go` 里这一段：
```go
// LoadConfig 读取磁盘上保存的任务参数；不存在则返回默认值。
func (a *App) LoadConfig() config.Config {
	path, err := config.DefaultPath()
	if err != nil {
		c := config.Config{}
		c.Normalize()
		return c
	}
	return config.Load(path)
}

// SaveConfig 把任务参数写入磁盘；出错时返回错误文案，成功返回空字符串。
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
```

替换成：
```go
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
```

- [ ] **Step 2: 编译确认通过**

Run: `cd ~/GolandProjects/githubbaidu && go build ./...`
Expected: 编译成功。如果 `frontend/dist` 还没构建过，`go:embed all:frontend/dist` 会因为目录为空而报错——这种情况下先跑一次 `cd frontend && npm install && npm run build`（不依赖本次改动，属于项目原有的已知前置条件，README 里也提到过）。

- [ ] **Step 3: 全量跑 go vet + go test**

Run: `cd ~/GolandProjects/githubbaidu && go vet ./... && go test ./...`
Expected: `go vet` 无输出；`internal/account`、`internal/article`、`internal/config` 全部 `ok`（`internal/accountpublish`、`internal/runlog` 因为测试文件已删除，`go test` 会显示 `?  githubbaidu/internal/accountpublish [no test files]` 这类提示，这是预期行为，不是失败）。

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add app.go
git commit -m "refactor: app.go 配置读写改用 dfclientkit appconfig 简化路径处理"
```

Go 侧迁移到此结束。此时应用的运行行为应该和迁移前完全一致——如果手头有时间，可以先跑一次 `wails dev` 验证 Go 侧改动没有引入回归，再继续 Phase D 前端改造。

---

## Phase D — githubbaidu 前端接入 df-ui-shell

### Task 17: package.json + vite.config.js 接入 df-ui-shell

`ui-shell` 是纯源码分发的包（`.vue` 文件不预编译），Vite 默认会把 `node_modules` 里的依赖当成"外部库"走 esbuild 预打包（`optimizeDeps`），但 esbuild 不认识 `.vue` 语法，所以要把这个包排除出预打包列表；另外 Vite 的开发服务器默认只允许访问项目根目录内的文件（`server.fs.allow`），而 `dfclientkit` 在githubbaidu 目录树之外，所以要显式放行。

**Files:**
- Modify: `githubbaidu/frontend/package.json`
- Modify: `githubbaidu/frontend/vite.config.js`

- [ ] **Step 1: 加依赖**

把 `frontend/package.json` 的 `dependencies` 块：
```json
  "dependencies": {
    "vue": "^3.5.0"
  },
```

改成：
```json
  "dependencies": {
    "vue": "^3.5.0",
    "@dongfang/df-ui-shell": "file:../../dfclientkit/ui-shell"
  },
```

- [ ] **Step 2: 跑 npm install，确认链接方式**

Run: `cd ~/GolandProjects/githubbaidu/frontend && npm install`
Expected: 安装成功。然后跑 `ls -la node_modules/@dongfang/df-ui-shell` 看一下这是个符号链接（`lrwxr-xr-x ... df-ui-shell -> ../../../dfclientkit/ui-shell`）还是被复制进去的真实目录——现代 npm（v7+）对 `file:` 协议依赖默认是建符号链接，两种情况下面 Step 3 的 `vite.config.js` 配置都覆盖到了，不用因为链接方式不同而改配置。

- [ ] **Step 3: 改 vite.config.js**

`githubbaidu/frontend/vite.config.js`（完整替换原内容）:
```js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const rootDir = dirname(fileURLToPath(import.meta.url)) // .../githubbaidu/frontend
const dfclientkitDir = resolve(rootDir, '../../dfclientkit')

export default defineConfig({
  plugins: [vue()],
  optimizeDeps: {
    // df-ui-shell 是纯源码分发的包（.vue 文件不预编译），
    // esbuild 的依赖预打包不认识 .vue 语法，必须排除掉。
    exclude: ['@dongfang/df-ui-shell'],
  },
  server: {
    fs: {
      // dfclientkit 在项目目录树之外（本地 file: 依赖），Vite 默认的
      // 文件系统访问白名单只到项目根目录，这里需要显式放行。
      allow: [rootDir, dfclientkitDir],
    },
  },
})
```

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add frontend/package.json frontend/package-lock.json frontend/vite.config.js
git commit -m "chore: 前端接入 @dongfang/df-ui-shell（本地 file: 依赖）"
```

---

### Task 18: NavRail 图标拆成独立小组件

NavRail 组件改造后不再自带图标，githubbaidu 需要把原来写死在 NavRail 里的 8 个 SVG（7 个主菜单 + 1 个"使用说明"）拆成独立的小组件传进去。

**Files:**
- Create: `githubbaidu/frontend/src/icons/PublishIcon.vue`
- Create: `githubbaidu/frontend/src/icons/ContentIcon.vue`
- Create: `githubbaidu/frontend/src/icons/ProxyIcon.vue`
- Create: `githubbaidu/frontend/src/icons/CaptchaIcon.vue`
- Create: `githubbaidu/frontend/src/icons/SpiderIcon.vue`
- Create: `githubbaidu/frontend/src/icons/LinksIcon.vue`
- Create: `githubbaidu/frontend/src/icons/CollectIcon.vue`
- Create: `githubbaidu/frontend/src/icons/HelpIcon.vue`

- [ ] **Step 1: 创建 8 个图标文件**

`githubbaidu/frontend/src/icons/PublishIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
    <path d="M8 11V3.5M8 3.5L5.2 6.3M8 3.5L10.8 6.3M2.5 10.5v2a1 1 0 001 1h9a1 1 0 001-1v-2" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/ContentIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
    <rect x="3.2" y="2.5" width="9.6" height="11" rx="1.2" />
    <path d="M5.5 6h5M5.5 8.3h5M5.5 10.6h3" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/ProxyIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
    <circle cx="8" cy="8" r="5.5" />
    <path d="M2.5 8h11M8 2.5c1.8 1.6 1.8 9.4 0 11M8 2.5c-1.8 1.6-1.8 9.4 0 11" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/CaptchaIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
    <path d="M8 2.3l4.7 1.7v3.6c0 3.2-2 5.7-4.7 6.6-2.7-.9-4.7-3.4-4.7-6.6V4L8 2.3z" stroke-linejoin="round" />
    <path d="M6 8.1l1.4 1.4L10.2 6.5" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/SpiderIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
    <path d="M8 2.5v2.4M8 11.1v2.4M2.5 8h2.4M11.1 8h2.4M4.3 4.3l1.7 1.7M10 10l1.7 1.7M11.7 4.3L10 6M6 10l-1.7 1.7" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/LinksIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
    <path d="M3 5h10M3 8h7M3 11h4" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/CollectIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
    <path d="M8 3.5V11M8 11L5.2 8.2M8 11l2.8-2.8M2.5 12.5v1a1 1 0 001 1h9a1 1 0 001-1v-1" />
  </svg>
</template>
```

`githubbaidu/frontend/src/icons/HelpIcon.vue`:
```vue
<template>
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
    <circle cx="8" cy="8" r="5.5" />
    <path d="M6.5 6.3c0-.9.7-1.5 1.5-1.5s1.5.6 1.5 1.5c0 .8-.7 1.1-1.1 1.4-.3.2-.4.5-.4.9" />
    <circle cx="8" cy="11" r=".7" fill="currentColor" stroke="none" />
  </svg>
</template>
```

- [ ] **Step 2: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add frontend/src/icons
git commit -m "feat: 拆分 NavRail 图标为独立小组件，供 App.vue 传给共享 NavRail"
```

---

### Task 19: App.vue 接入共享组件

**Files:**
- Modify: `githubbaidu/frontend/src/App.vue`

- [ ] **Step 1: 改 import 块**

把 `App.vue` 顶部这一段：
```js
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import TitleBar from './components/TitleBar.vue'
import NavRail from './components/NavRail.vue'
import PublishPage from './components/PublishPage.vue'
import HelpPage from './components/HelpPage.vue'
import PlaceholderPage from './components/PlaceholderPage.vue'
import RunParamsPanel from './components/RunParamsPanel.vue'
import LogPanel from './components/LogPanel.vue'
import StatusBar from './components/StatusBar.vue'
import ResultsModal from './components/ResultsModal.vue'
import * as App from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

const APP_VERSION = 'v1.0.0'
```

替换成：
```js
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { TitleBar, NavRail, StatusBar, LogPanel, ResultsModal } from '@dongfang/df-ui-shell'
import PublishPage from './components/PublishPage.vue'
import HelpPage from './components/HelpPage.vue'
import PlaceholderPage from './components/PlaceholderPage.vue'
import RunParamsPanel from './components/RunParamsPanel.vue'
import PublishIcon from './icons/PublishIcon.vue'
import ContentIcon from './icons/ContentIcon.vue'
import ProxyIcon from './icons/ProxyIcon.vue'
import CaptchaIcon from './icons/CaptchaIcon.vue'
import SpiderIcon from './icons/SpiderIcon.vue'
import LinksIcon from './icons/LinksIcon.vue'
import CollectIcon from './icons/CollectIcon.vue'
import HelpIcon from './icons/HelpIcon.vue'
import * as App from '../wailsjs/go/main/App'
import { EventsOn, WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime/runtime'

const APP_VERSION = 'v1.0.0'
const APP_NAME = 'GitHub 文章发布器'

const NAV_ITEMS = [
  { key: 'publish', cn: '发布', en: 'PUBLISH', icon: PublishIcon },
  { key: 'content', cn: '内容设置', en: 'CONTENT', icon: ContentIcon },
  { key: 'proxy', cn: 'IP 设置', en: 'PROXY', icon: ProxyIcon },
  { key: 'captcha', cn: '打码设置', en: 'CAPTCHA', icon: CaptchaIcon },
  { key: 'spider', cn: '蜘蛛设置', en: 'SPIDER', icon: SpiderIcon },
  { key: 'links', cn: '其他设置', en: 'URL / LINKS', icon: LinksIcon },
  { key: 'collect', cn: '采集文章', en: 'COLLECT', icon: CollectIcon },
]
const NAV_BOTTOM_ITEMS = [{ key: 'help', cn: '使用说明', en: 'HELP', icon: HelpIcon }]
```

- [ ] **Step 2: 改 TitleBar 用法**

把模板里的：
```html
    <TitleBar :theme="theme" :version="APP_VERSION" :pill-text="pill.text" :pill-kind="pill.kind" @toggle-theme="toggleTheme" />
```

替换成：
```html
    <TitleBar
      :theme="theme"
      :version="APP_VERSION"
      :pill-text="pill.text"
      :pill-kind="pill.kind"
      :app-name="APP_NAME"
      logo-text="G"
      @toggle-theme="toggleTheme"
      @minimize="WindowMinimise"
      @maximize="WindowToggleMaximise"
      @close="Quit"
    />
```

- [ ] **Step 3: 改 NavRail 用法**

把模板里的：
```html
      <NavRail :page="page" @navigate="(p) => (page = p)" />
```

替换成：
```html
      <NavRail :page="page" :items="NAV_ITEMS" :bottom-items="NAV_BOTTOM_ITEMS" @navigate="(p) => (page = p)" />
```

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add frontend/src/App.vue
git commit -m "refactor: App.vue 接入 df-ui-shell 的 TitleBar/NavRail/StatusBar/LogPanel/ResultsModal"
```

---

### Task 20: RunParamsPanel 的 NumberStepper 引用改到共享包

**Files:**
- Modify: `githubbaidu/frontend/src/components/RunParamsPanel.vue:2`

- [ ] **Step 1: 改 import**

把 `RunParamsPanel.vue` 第 2 行：
```js
import NumberStepper from './NumberStepper.vue'
```

替换成：
```js
import { NumberStepper } from '@dongfang/df-ui-shell'
```

- [ ] **Step 2: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add frontend/src/components/RunParamsPanel.vue
git commit -m "refactor: RunParamsPanel 的 NumberStepper 改为从 df-ui-shell 引入"
```

---

### Task 21: 清理已被替换的本地文件

**Files:**
- Delete: `githubbaidu/frontend/src/components/TitleBar.vue`
- Delete: `githubbaidu/frontend/src/components/NavRail.vue`
- Delete: `githubbaidu/frontend/src/components/StatusBar.vue`
- Delete: `githubbaidu/frontend/src/components/LogPanel.vue`
- Delete: `githubbaidu/frontend/src/components/NumberStepper.vue`
- Delete: `githubbaidu/frontend/src/components/ResultsModal.vue`
- Delete: `githubbaidu/frontend/src/styles/theme.css`
- Modify: `githubbaidu/frontend/src/main.js`

- [ ] **Step 1: 删除已经被 df-ui-shell 取代的本地组件与主题文件**

Run:
```bash
cd ~/GolandProjects/githubbaidu/frontend/src
rm components/TitleBar.vue components/NavRail.vue components/StatusBar.vue \
   components/LogPanel.vue components/NumberStepper.vue components/ResultsModal.vue \
   styles/theme.css
rmdir styles 2>/dev/null || true
```

- [ ] **Step 2: 改 main.js，主题改成引入共享包的**

`githubbaidu/frontend/src/main.js`（完整替换原内容）:
```js
import { createApp } from 'vue'
import App from './App.vue'
import '@dongfang/df-ui-shell/theme.css'

createApp(App).mount('#app')
```

- [ ] **Step 3: 全量确认没有遗留引用**

Run: `cd ~/GolandProjects/githubbaidu/frontend && grep -rn "styles/theme.css\|components/TitleBar\|components/NavRail\|components/StatusBar\|components/LogPanel\|components/NumberStepper\|components/ResultsModal" src/`
Expected: 无输出（没有任何文件还在引用这几个被删除的本地路径）。

- [ ] **Step 4: 提交**

```bash
cd ~/GolandProjects/githubbaidu
git add -A frontend/src
git commit -m "chore: 删除已迁移到 df-ui-shell 的本地组件与主题文件"
```

---

### Task 22: 端到端验证

**Files:** 无代码改动，纯验证。

- [ ] **Step 1: 前端能正常构建**

Run: `cd ~/GolandProjects/githubbaidu/frontend && npm run build`
Expected: 构建成功，`frontend/dist` 生成产物，没有 "Cannot find module '@dongfang/df-ui-shell'" 或 SFC 编译错误一类的报错。如果报 "@dongfang/df-ui-shell" 找不到，先确认 Task 17 的 `npm install` 真的把符号链接建出来了（`ls -la node_modules/@dongfang`），必要时重新 `npm install`。

- [ ] **Step 2: Go 侧整体编译与测试**

Run: `cd ~/GolandProjects/githubbaidu && go build ./... && go vet ./... && go test ./...`
Expected: 全部通过。

- [ ] **Step 3: wails dev 手动过一遍完整交互**

Run: `cd ~/GolandProjects/githubbaidu && wails dev`

手动检查这些点，确认和迁移前行为、视觉一致：
- 标题栏：应用名"GitHub 文章发布器"、版本号、状态胶囊正常显示；明暗主题切换正常；最小化/最大化/关闭按钮都能正常控制窗口。
- 左侧导航：7 个菜单项图标和文案都在，点击能正常切换页面；底部"使用说明"按钮正常。
- 发布页：导入账号（文件/粘贴剪贴板）、移出账号、标记坏号、单独测试账号、清空账号、导出结果 CSV，全部正常。
- 右侧运行参数面板：数字输入框（线程数/间隔秒数/每号次数/失败换号次数/循环轮数/轮间隔/关键词插槽）都能编辑；选择文件夹/添加文件导入内容；开始/暂停/恢复/停止发布全流程正常，进度、轮次、日志实时更新。
- 底部日志面板：日志正常滚动、复制、导出、清空都正常。
- 底部状态栏：账号/成功/失败/待发/已用时数字正常更新。
- 查看结果弹窗：能正常打开、显示结果列表（CK/标题/结果三列）、复制全部、关闭。

- [ ] **Step 4: 记录验证结果**

如果以上全部通过，Go 侧和前端侧的迁移都已完成，githubbaidu 现在完全跑在 dfclientkit 基建之上，且和迁移前行为一致。如果发现任何回归，回到对应 Task 定位问题（不要跳过 Task 顺序去改后面的文件）。
