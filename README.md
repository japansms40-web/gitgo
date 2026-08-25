# Git MD

本地 Markdown 内容生成器：按模板和词库批量生成文稿草稿，预览后导出成 `.md` 文件。带图形界面，单可执行文件，全程离线，不联网也不发布到任何地方。

## 使用
1. 「内容设置 → 模板」里写标题模板和正文模板，用占位符标出每篇要变的地方，其余文字原样保留。
2. 「词库」页里填关键词和变量库，一行一条。关键词可选顺序轮转或随机抽取。
3. 右侧设好生成篇数，点「生成」，在「预览」页逐条展开查看结果。
4. 点「导出 .md」选一个文件夹，每篇存成一个 Markdown 文件，标题作为文件名。

## 占位符
| 占位符 | 含义 |
|---|---|
| `{关键词}` | 关键词库里的当前一条 |
| `{变量1}` … `{变量5}` | 从对应变量库随机抽一行 |
| `{英文=6}` / `{小写=4}` / `{数字=3}` | 指定位数的随机串 |
| `{日期}` / `{时间}` | 生成时刻，如 `2026-08-25` / `17:42` |

同一个占位符出现多次会各自独立重新抽取；写错的占位符原样留在结果里，方便发现笔误。

## 说明
- 模板和词库以 txt 文件存在本机（macOS: `~/Library/Application Support/gitmd/content/`），每个输入框对应一个文件，可以直接用文本编辑器批量编辑——点「打开素材目录」即可跳转。
- 生成参数（篇数、关键词处理方式、去重开关等）保存在同目录的 `config.json`。
- 可选的正文后处理：去除重复行、仅保留含中文的行。空行不参与，段落结构会保留。

## 技术栈
Go 后端（生成引擎见 `internal/contentgen`，txt 读写见 `internal/contentstore`）+ Vue 3 前端（`frontend/`），通过 [Wails](https://wails.io) 打包成单个原生桌面可执行文件。

## 开发
- 首次运行前装好 Node.js 和 `wails` CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）。
- 热重载调试：`wails dev`
- `frontend/dist` 未纳入版本控制，`go build`/`go vet`/`go test` 依赖它先被构建出来（哪怕只是 `wails dev` 跑一次或 `cd frontend && npm install && npm run build`），否则 `main.go` 里的 `//go:embed all:frontend/dist` 会因为目录为空而编译失败。

## 构建
- macOS: `make -f build/Makefile mac` → `build/bin/gitmd.app`
- Windows: `make -f build/Makefile windows` → `build/bin/gitmd.exe`
