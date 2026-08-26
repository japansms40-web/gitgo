# GitHub 仓库列表客户端（GET /repos）设计

日期：2026-08-26
状态：已确认，待实现

## 背景与目标

`docs/flows_1_requests/01_GET_repos.txt` 抓的是 GitHub **网页版**（`github.com`，非 `api.github.com`）
内部 JSON 接口 `GET /repos?q=owner:@me&page=N`，靠浏览器 **Cookie 会话**鉴权。目标是在 gitgo 里
用项目统一的 gohttpkit 基建（`internal/httpclient`）复刻这条请求，端点函数写法参考
insgo 的 `android/v361/followers.go`（构参 → 发请求 → 校状态码 → 解响应）。

本次范围（已与用户确认）：**只实现 GET /repos 这一个接口**，但把 GitHub 客户端包骨架搭好，
后续接口（Copilot 上下文、创建仓库等）可直接往里加。凭证阶段**先写死**，用测试用例验证。

## 非目标（YAGNI）

- 不做多账号凭证管理 / Cookie 轮转 / 会话回写。
- 不接前端 Wails 绑定（本次只到 Go 客户端 + 测试）。
- 不实现 Copilot、创建仓库等其它抓包里的接口。
- 不走官方 PAT + `api.github.com`（那是另一套接口，不复刻本抓包）。

## 架构

新增包 `internal/github`，薄 `Client` 结构体包住 `*httpx.Client`，端点方法挂在其上
（对齐 insgo `pkg.Client` 的扩展思路）。HTTP 一律经 `internal/httpclient.New(...)` 构造，
不裸用 `net/http`。

```
internal/github/
  client.go          # github.Client 构造：BaseURL=https://github.com + 复刻网页头 + Cookie 入参
  repos.go           # ListRepos(ctx, page) 端点函数（仿 followers.go）
  repos_types.go     # 响应结构体（对着抓包 JSON）
  testdata/
    repos_response.json  # 从 01_GET_repos.txt 抠出的 RESPONSE 响应体，供离线测试
  repos_test.go      # 离线解析测试 + 集成测试
```

## 组件

### client.go — `github.Client`

- `type Client struct { http *httpx.Client }`
- `func New(cookie string) (*Client, error)`：
  - 用 `httpclient.New(httpclient.Config{ BaseURL: "https://github.com", Headers: {...} })` 建底层
    `*httpx.Client`。
  - `Headers`（key 全小写）复刻抓包关键头：
    - `accept: application/json`
    - `content-type: application/json`
    - `x-requested-with: XMLHttpRequest`
    - `github-is-react: true`
    - `github-verified-fetch: true`
    - `x-github-app-type: dataRouter`
    - `user-agent: <抓包同款 Chrome UA>`
    - `accept-language: zh-CN,zh;q=0.9`
    - `cookie: <入参 cookie>`
  - Cookie 作为**入参**传入，不写死进生产代码（避免明文 Cookie 落库；见备忘 github-token-plaintext-risk）。
- `New(cookie string, opts ...Option)` 用函数式选项留扩展点。`WithProxy(url)` 设代理
  （socks5/http/https）。**注意 gohttpkit 空 ProxyURL 时不回退环境变量、直接直连**，而 github.com
  常需代理，故代理必须显式传。
- 说明：`Connection` / `Accept-Encoding` / `Host` 等 hop-by-hop 或由 gohttpkit/传输层自动管理的头
  不手动带。

### repos.go — `ListRepos`

- `func (c *Client) ListRepos(ctx context.Context, page int) (*ReposResponse, error)`
- 构 `url.Values{ "q": {"owner:@me"}, "page": {strconv.Itoa(page)} }`（page<1 归一到 1）。
- `body, err := c.http.Get(ctx, "/repos", params)`。
- 状态码校验：`code := c.http.SnapshotResponseStatusCode()`；非 2xx（如 Cookie 过期返回的 401/403）
  返回带状态码与响应体片段的 error（默认链不替你判非 2xx）。
- `json.Unmarshal(body, &resp)`，返回 `*ReposResponse`。

### repos_types.go — 响应结构体

对着抓包 JSON：

```json
{"meta":{"title":"Repositories"},"payload":{"reposFinderPageRoute":{
  "pageCount":1,
  "repositories":[{"type":"Public","name":"abyuds","owner":"VKMW57by18","isFork":false,
    "description":"","allTopics":[],"primaryLanguage":null,"pullRequestCount":0,"issueCount":0,
    "starsCount":0,"forksCount":0,"license":null,"participation":null,
    "lastUpdated":{"hasBeenPushedTo":true,"timestamp":"2026-08-23T16:10:15.134Z"},
    "isAdminableByUser":true}],
  "repositoryCount":1,"additionalFilterProviders":[]}}}
```

映射：
- `ReposResponse{ Meta struct{Title string}; Payload Payload }`
- `Payload{ ReposFinderPageRoute ReposFinderPageRoute `json:"reposFinderPageRoute"` }`
- `ReposFinderPageRoute{ PageCount int; Repositories []Repository; RepositoryCount int }`
- `Repository{ Type,Name,Owner,Description string; IsFork,IsAdminableByUser bool;
   AllTopics []string; PullRequestCount,IssueCount,StarsCount,ForksCount int;
   LastUpdated LastUpdated; PrimaryLanguage,License,Participation json.RawMessage }`
   - `primaryLanguage/license/participation` 抓包为 null、形状不定，用 `json.RawMessage` 兜底。
- `LastUpdated{ HasBeenPushedTo bool; Timestamp string }`

### 测试（两个都要）

1. **离线解析测试** `TestListRepos_DecodeResponse`：读 `testdata/repos_response.json`，`json.Unmarshal`，
   断言 `RepositoryCount==1`、`PageCount==1`、`Repositories[0].Name=="abyuds"`、`Type=="Public"`、
   `LastUpdated.HasBeenPushedTo==true`。永远稳定、不依赖网络。
2. **集成测试** `TestListRepos_Live`：`if testing.Short() { t.Skip(...) }` 守卫；Cookie 从环境变量
   `GH_TEST_COOKIE` 读（**绝不写死进仓库**——本仓库公开，Cookie 泄露即会话被劫持），未设则 `t.Skip`；
   代理取 `GITHUB_TEST_PROXY` 或系统 `HTTPS_PROXY/HTTP_PROXY`，`New(cookie, WithProxy(...))` →
   `ListRepos(ctx, 1)`，断言无 error 且 `PageCount>=1`。跑法：
   `GH_TEST_COOKIE='...' go test -run Live ./internal/github/`。`go test -short` 只跑离线。
   Cookie 过期时测试会因非 2xx 报错，届时换环境变量那串即可。

## 数据流

`Client.ListRepos(ctx, page)` → httpx 默认链（trace 日志 / 四种解压 / 网络层重试+退避 / 状态码与
响应头缓存）→ github.com → `[]byte` → 状态码校验 → `json.Unmarshal` → `*ReposResponse`。

## 错误处理

- 构头失败 / 网络错误：由 httpx 直接返回 error（网络类错误会按默认策略重试 3 次）。
- 非 2xx：端点函数显式判 `SnapshotResponseStatusCode()`，返回形如
  `github: list repos 返回非 2xx 状态码 401` 的 error（附响应体前若干字节便于排查）。
- JSON 解析失败：包装后返回。

## 验收标准

- `go test -short ./internal/github/` 通过（离线测试绿）。
- `go test ./internal/github/`（Cookie 未过期时）集成测试也通过，能真实取到仓库列表。
- 全项目 `go build ./...` 通过，不引入 `net/http` 裸用。
