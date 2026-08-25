# GitHub 文章发布器

用账号队列（CK/UA/IP）把本地 Markdown 内容批量发布到自有系统。带图形界面，单可执行文件。

## 使用
1. 右侧运行参数面板「选择文件夹」或「添加文件」导入 `.md`/`.txt` 作为待发布内容（文件名作为标题）。
2. 发布页导入账号：点「导入账号」选 TXT 文件，或双击「双击粘贴剪贴板」粘贴，多个账号用 `----` 分隔。
3. 可设发布间隔与失败重试次数，点「开始发布」。账号队列逐个发布一篇内容（内容不够会循环使用），进度、日志实时显示，可「停止」。
4. 右键账号行可复制 CK、单独测试、移出列表或标记为坏号（批量发布时会跳过坏号）；顶部可搜索、按状态筛选、导出结果。

## 说明
- 账号队列长期持久化，成功/失败/总数是跨多次发布的累计值。
- 项目尚未接入目标系统的真实发布协议（CK 怎么带、UA/IP 怎么用、怎么判定成功失败），发布/测试目前走占位实现（`internal/accountpublish.TODORequester`），会直接返回失败，方便先验证队列与状态流转；接口细节确定后实现一个新的 `Requester` 换掉它即可。
- 运行参数保存在本机的配置文件中（macOS: `~/Library/Application Support/ghpublisher/config.json`），账号队列保存在同目录的 `accounts.json`。

## 技术栈
Go 后端（业务逻辑见 `internal/`） + Vue 3 前端（`frontend/`），通过 [Wails](https://wails.io) 打包成单个原生桌面可执行文件。

## 开发
- 首次运行前装好 Node.js 和 `wails` CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）。
- 热重载调试：`wails dev`
- `frontend/dist` 未纳入版本控制，`go build`/`go vet`/`go test` 依赖它先被构建出来（哪怕只是 `wails dev` 跑一次或 `cd frontend && npm install && npm run build`），否则 `main.go` 里的 `//go:embed all:frontend/dist` 会因为目录为空而编译失败。

## 构建
- macOS: `make -f build/Makefile mac` → `build/bin/ghpublisher.app`
- Windows: `make -f build/Makefile windows` → `build/bin/ghpublisher.exe`
