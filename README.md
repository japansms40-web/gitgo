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
- 设置（含 Token）保存在系统的 Fyne Preferences 中。

## 构建
- macOS: `make -f build/Makefile mac` → `dist/ghpublisher`
- Windows: `make -f build/Makefile windows` → `dist/ghpublisher.exe`
