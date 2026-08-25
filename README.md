# GitHub 文章发布器

用个人 Access Token 把本地 Markdown 文章批量发布到你自己的 GitHub 仓库。带图形界面，单可执行文件。

## 使用
1. 在 GitHub 生成 Personal Access Token（勾选 `repo` 权限；若要自动建仓库还需 `public_repo`）。
2. 打开程序，填入 Token、Owner、Repo、分支、目标目录，点「验证 Token」。
3. 「选择文件夹」或「添加文件」导入 `.md`/`.txt`（`.txt` 会以 `.md` 提交，文件名作为标题）。
4. 点「开始发布」。进度、状态、结果链接实时显示，可「停止」，可「导出链接列表」。

## 说明
- 单账号、单仓库；文件已存在则更新覆盖。
- 每篇一次提交，可设发布间隔与失败重试；遇 API 限流会按 `Retry-After` 等待重试。
- 设置（含 Token）保存在本机的配置文件中（macOS: `~/Library/Application Support/ghpublisher/config.json`）。

## 技术栈
Go 后端（业务逻辑见 `internal/`） + Vue 3 前端（`frontend/`），通过 [Wails](https://wails.io) 打包成单个原生桌面可执行文件。

## 开发
- 首次运行前装好 Node.js 和 `wails` CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）。
- 热重载调试：`wails dev`
- `frontend/dist` 未纳入版本控制，`go build`/`go vet`/`go test` 依赖它先被构建出来（哪怕只是 `wails dev` 跑一次或 `cd frontend && npm install && npm run build`），否则 `main.go` 里的 `//go:embed all:frontend/dist` 会因为目录为空而编译失败。

## 构建
- macOS: `make -f build/Makefile mac` → `build/bin/ghpublisher.app`
- Windows: `make -f build/Makefile windows` → `build/bin/ghpublisher.exe`
