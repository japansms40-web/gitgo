# dfclientkit：多客户端项目共用基建 设计

## 背景

githubbaidu（`ghpublisher`，GitHub 文章发布器）是第一个基于 Wails + Vue3 打包成桌面客户端的"账号队列驱动发布器"。以后会有多个同类项目（不同目标平台的发布器），每个项目独立打包、独立运行，但账号队列管理、并发任务引擎、配置持久化、运行日志格式化、桌面壳 UI 这些能力是重复的。

参考了两个项目的基建思路：

- **insgo**：私有 Go 模块（`pkg/`、`internal/`、顶层 `auth/`/`logger/`/`manifest/`），被多个 `cmd/`、`web/` 业务代码通过 `go get` + `replace` 消费，一套底层能力支撑多个独立入口。
- **AppAutoToolsDf**：单仓库内的分层基建（`core` 是叶子包、供应商包互不依赖、组合根统一装配），并用 `ArchitectureTest` 靠 import 扫描强制依赖方向，而不是靠自觉遵守。

本设计把 githubbaidu 里已经足够通用的部分（账号队列、并发发布引擎、配置持久化、日志格式化、桌面壳组件）抽成独立的共享基建 `dfclientkit`，githubbaidu 改为消费它，作为新项目起步的参照实现。

## 目标 / 非目标

**目标**：
- 新建独立仓库 `dfclientkit`，承载 Go 后端基建与前端 UI 壳基建。
- githubbaidu 迁移为消费 `dfclientkit`，功能与界面保持不变（这是一次纯基建搬迁，不新增/改变业务行为）。
- 以后新项目能以"新建仓库 + 两条本地依赖 + 写自己的业务页面和 `Requester` 实现"的方式起步。

**非目标**：
- 不在本次设计范围内新增任何新平台的发布器。
- 不引入 Go 泛型化的 `Account[T]`（身份字段目前统一是 CK/UA/IP 形状，按现状固定字段更简单；以后遇到明显不同形状的账号身份，再单独扩展）。
- 不为 `dfclientkit` 额外写 `ArchitectureTest` 式的依赖扫描测试——它是独立仓库，Go module 系统本身已经物理阻止了"基建反向依赖业务项目"，这层强制力由仓库边界提供，不需要重复造轮子。
- 不搭建私有 npm/Go module registry。本地开发用 `replace`（Go）/ `file:`（npm）协议指向本地兄弟目录；若未来需要真正独立分发，再按 insgo 的模式（module path 换成实际 git 仓库地址）追加。

## 仓库与依赖拓扑

新建独立仓库 `~/GolandProjects/dfclientkit`，与 `githubbaidu`、`insgo` 同级：

```
dfclientkit/
  go/                          Go 模块（module dfclientkit）
    account/                   账号队列模型
    taskrunner/                并发任务引擎
    appconfig/                 通用配置持久化
    runlog/                    事件 -> 日志文案
    go.mod
  ui-shell/                    npm 包 @dongfang/df-ui-shell
    src/
      TitleBar.vue
      NavRail.vue
      StatusBar.vue
      LogPanel.vue
      NumberStepper.vue
      ResultsModal.vue
      theme.css
    package.json
  README.md
```

githubbaidu 侧的引用方式：

- `go.mod`：新增 `require dfclientkit v0.0.0`，并用 `replace dfclientkit => ../dfclientkit/go` 指向本地目录（与 insgo 消费 `insgouagen` 的本地开发模式一致）。
- `frontend/package.json`：新增 `"@dongfang/df-ui-shell": "file:../../dfclientkit/ui-shell"`，npm 的 `file:` 协议是 Go 本地 `replace` 的等价物，不需要搭私有 registry。

依赖方向靠仓库物理边界锁死：`dfclientkit` 是独立仓库，不会有人在它的 `go.mod`/`package.json` 里意外反向依赖 `githubbaidu`——这就是 AppAutoToolsDf 用 `ArchitectureTest` 想达到的效果，这里由仓库边界天然提供，不需要额外写检测代码。

以后新项目起步方式：新建仓库 → 两条本地依赖指过来 → 用 `taskrunner` 的引擎 + `df-ui-shell` 的壳，只写自己的业务页面和 `Requester` 实现。

## dfclientkit/go：四个包

### account

`Account{CK, UA, IP, Status, Success, Fail, Total, Bad}` 原样搬过来，状态常量 `StatusPending`/`StatusRunning`/`StatusSuccess`/`StatusFailed`、`ParseImportText`（按 `----` 分隔解析）一起搬。

`Load` 必须保证永远返回非 nil 切片——这是 githubbaidu 刚修过的坑（`commit 493fb57`：账号/结果列表返回裸 nil 切片，序列化成 JSON 变成 `null`，前端 `.length`/`.map` 直接崩页面）。这条注释和实现细节要带过去，避免后续项目重新踩一遍。

### taskrunner

原 `accountpublish` 包搬过来并改名（去掉 GitHub/发布语境）：

- `RunConfig{Threads, IntervalSec, PerAccountCount, FailSwitchCount, CycleRounds, RoundIntervalSec}` + `normalize()`
- `PauseGate`（暂停/恢复/取消）
- `Requester`/`RepoCreator` 接口——真正的协议由各项目注入，`taskrunner` 本身不关心目标平台是什么
- `Runner.Run(...)`：并发处理账号池、单账号循环发布、连续失败换号、多轮循环、暂停/恢复/取消、事件回调，整段逻辑搬过来
- `Event`：`ArticleTitle` 字段改名为 `ItemLabel`——不是所有发布器都发"文章"，这个字段名不该带着文章语境泄漏到引擎层

`TODORequester`/`TODORepoCreator` 这类项目专属的占位实现**不带走**，留在 githubbaidu（它们是"GitHub 真实协议还没接入"这个具体项目状态的产物，不是基建）。

### appconfig

把"读 JSON 配置、按系统目录存、缺省兜底"这套机制抽成泛型工具：

```go
func DefaultDir(appName string) (string, error)
func Load[T any](appName string, defaults T) T
func Save[T any](appName string, cfg T) error
```

每个项目自己的 `Config` 结构体内嵌 `taskrunner.RunConfig` 加项目专属字段（githubbaidu 就是 `KeywordSlots`、`CreateRepo`），调用方式：

```go
cfg := appconfig.Load[Config]("ghpublisher", defaults())
```

### runlog

`TagFor`/`LineFor` 基于 `taskrunner.EventKind` 生成日志标签与文案。文案去掉"发布"这个专属动词，缺省用"处理"，但保留一个可选的 `verb string` 参数——调用方想显示"发布"/"注册"/"采集"就传进去，不传就是"处理"。这样引擎层日志不会绑死在某一种具体业务动作上。

## df-ui-shell：六个组件 + 主题

现状里已经是纯 props/emit 驱动、没有业务耦合，**原样搬**：

- `StatusBar.vue`（total/success/fail/pending/elapsed）
- `LogPanel.vue`（lines/autoScroll，复制/导出/清空）
- `NumberStepper.vue`（label/unit/modelValue/editable/min）

以下三个需要先去掉硬编码才能通用：

- **TitleBar**：`GitHub 文章发布器` 应用名和 logo 字母 `G` 改成 `appName`/`logoText` prop；主题的 `localStorage` 读写交还给调用方管理，组件本身不再直接读写 `localStorage`。窗口控制按钮（最小化/最大化/关闭）继续走 `wailsjs/runtime`，因为每个消费方本身就是 Wails 项目，自带这个依赖。
- **NavRail**：现在写死了 7 个固定菜单项（发布/内容设置/IP设置/打码设置/蜘蛛设置/其他设置/采集文章）和对应 SVG。改成 `items: [{key, cn, en, icon?}]` 从外部传入，`icon` 是可选的 Vue 组件，`NavRail` 用 `<component :is="item.icon" />` 渲染；不传 `icon` 就只显示文字。
- **ResultsModal**：结果列目前写死"CK/标题/结果"三列。改成 `columns: [{key, label}]` 可配置，默认值保持现在这三列，githubbaidu 现有调用代码不用改。

`theme.css` 的 CSS 变量（`--bg`/`--nav`/`--border`/`--accent`/`--text`/`--muted` 等）原样作为基础主题搬过去，各项目可以在自己的 CSS 里覆盖强调色等个性化变量。

`PublishPage.vue`、`RunParamsPanel.vue`、`HelpPage.vue`、`PlaceholderPage.vue` 因为耦合了 GitHub 发布器专属字段（`keywordSlots`、`createRepo`、内容文件夹扫描 UI 等），**留在 githubbaidu**，不进基建。

## githubbaidu 迁移

分两步，不一把梭：

1. **Go 侧**：`internal/account`、`internal/config`、`internal/accountpublish`、`internal/runlog` 改成薄薄一层委托/别名到 `dfclientkit`，`go test ./...` 全绿。
2. **前端侧**：六个组件换成从 `@dongfang/df-ui-shell` 引入，传入 `appName="GitHub 文章发布器"`、`navItems=[...]`（保留现在这 7 项及对应图标）等 props；`wails dev` 手动过一遍发布/暂停/停止/账号导入导出/查看结果全流程，确认界面和交互与改造前一致。

保留在 githubbaidu、不进基建的部分：`internal/article`（本地 Markdown/文本扫描）、`TODORequester`/`TODORepoCreator`、`Config` 里的 `KeywordSlots`/`CreateRepo` 字段。

## 验证方式

- `dfclientkit/go`：每个包（account/taskrunner/appconfig/runlog）自带单测，覆盖现有 githubbaidu 测试里已验证过的行为（账号导入解析、并发换号/轮次逻辑、暂停恢复、配置读写默认值兜底、日志文案）。
- githubbaidu Go 侧迁移后：`go vet ./...` + `go test ./...` 全绿。
- githubbaidu 前端迁移后：`wails dev` 手动跑一遍发布/暂停/停止/账号导入/粘贴/移出/标记坏号/导出结果/查看结果/日志复制导出清空/明暗主题切换，确认功能和视觉都没有回归。
