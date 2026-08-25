package contentstore

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitmd/internal/contentgen"
)

func TestLoadSeedsExamplesWhenMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "content")

	lib, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	if lib.TitleTemplate == "" {
		t.Error("首次运行应写入示例标题模板")
	}
	if len(lib.BodyTemplates) != contentgen.BodyTemplateCount {
		t.Fatalf("正文模板应有 %d 套，实际 %d 套", contentgen.BodyTemplateCount, len(lib.BodyTemplates))
	}
	for i, tpl := range lib.BodyTemplates {
		if strings.TrimSpace(tpl) == "" {
			t.Errorf("正文模板 %d 应有示例内容", i)
		}
	}
	if len(lib.Keywords) == 0 {
		t.Error("首次运行应写入示例关键词")
	}
	if len(lib.Vars) != contentgen.VarBankCount {
		t.Fatalf("变量库应有 %d 组，实际 %d 组", contentgen.VarBankCount, len(lib.Vars))
	}

	// 文件确实落到了磁盘上，用户能直接用文本编辑器打开。
	for _, path := range []string{
		filepath.Join(dir, titleFile), filepath.Join(dir, keywordsFile),
		bodyPath(dir, 0), bodyPath(dir, 1),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("应自动创建 %s: %v", filepath.Base(path), err)
		}
	}
	for i := 1; i <= contentgen.VarBankCount; i++ {
		if _, err := os.Stat(varPath(dir, i)); err != nil {
			t.Errorf("应自动创建变量%d.txt: %v", i, err)
		}
	}
}

// TestDefaultLibraryGenerates 保证内置示例本身是能跑通的：
// 装好就点「生成」不会报错，也不会留下没替换的占位符。
func TestDefaultLibraryGenerates(t *testing.T) {
	lib, err := Load(filepath.Join(t.TempDir(), "content"))
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	drafts, err := contentgen.Generate(lib, contentgen.Options{Count: 6}, rand.New(rand.NewSource(1)), time.Now())
	if err != nil {
		t.Fatalf("用默认素材生成失败: %v", err)
	}
	for i, d := range drafts {
		if strings.Contains(d.Title, "{") || strings.Contains(d.Body, "{") {
			t.Errorf("第 %d 篇残留了未替换的占位符:\n标题 %q\n正文 %q", i, d.Title, d.Body)
		}
		if strings.TrimSpace(d.Body) == "" {
			t.Errorf("第 %d 篇正文为空", i)
		}
	}
}

func TestLoadDoesNotClobberExistingFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, titleFile), []byte("已有内容"), 0o644); err != nil {
		t.Fatal(err)
	}

	lib, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}
	if lib.TitleTemplate != "已有内容" {
		t.Errorf("已存在的文件不应被覆盖，读到 %q", lib.TitleTemplate)
	}
}

func TestLoadKeepsClearedFileEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, contentgen.Library{}); err != nil {
		t.Fatal(err)
	}

	lib, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}
	if lib.TitleTemplate != "" || len(lib.Keywords) != 0 {
		t.Errorf("用户清空过的素材不应被示例内容重新填上，得到 %+v", lib)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := contentgen.Library{
		TitleTemplate: "{关键词}-{日期}",
		BodyTemplates: []string{"第一行\n\n第二行 {变量1}", "另一套 {变量2}"},
		Keywords:      []string{"茶叶", "咖啡"},
		Vars:          make([][]string, contentgen.VarBankCount),
	}
	want.Vars[0] = []string{"甲", "乙"}
	want.Vars[4] = []string{"末"}

	if err := Save(dir, want); err != nil {
		t.Fatalf("Save 返回错误: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	if got.TitleTemplate != want.TitleTemplate {
		t.Errorf("标题模板 = %q，期望 %q", got.TitleTemplate, want.TitleTemplate)
	}
	for i := range want.BodyTemplates {
		if got.BodyTemplates[i] != want.BodyTemplates[i] {
			t.Errorf("正文模板 %d = %q，期望 %q", i, got.BodyTemplates[i], want.BodyTemplates[i])
		}
	}
	if strings.Join(got.Keywords, ",") != strings.Join(want.Keywords, ",") {
		t.Errorf("关键词 = %v，期望 %v", got.Keywords, want.Keywords)
	}
	for i := range want.Vars {
		if strings.Join(got.Vars[i], ",") != strings.Join(want.Vars[i], ",") {
			t.Errorf("变量%d = %v，期望 %v", i+1, got.Vars[i], want.Vars[i])
		}
	}
}

// SaveTemplates 只写模板，必须保留用户在文件库里改过的关键词/图片/变量。
func TestSaveTemplatesKeepsDataFiles(t *testing.T) {
	dir := t.TempDir()
	// 先落一份带词库的完整素材。
	full := contentgen.Library{
		TitleTemplate: "旧标题",
		BodyTemplates: []string{"旧正文A", "旧正文B"},
		Keywords:      []string{"关键词甲", "关键词乙"},
		Images:        []string{"图一"},
		Vars:          make([][]string, contentgen.VarBankCount),
	}
	full.Vars[0] = []string{"变量甲"}
	if err := Save(dir, full); err != nil {
		t.Fatalf("Save 返回错误: %v", err)
	}

	// 只改模板。
	if err := SaveTemplates(dir, "新标题", []string{"新正文A", "新正文B"}); err != nil {
		t.Fatalf("SaveTemplates 返回错误: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}
	if got.TitleTemplate != "新标题" || got.BodyTemplates[0] != "新正文A" {
		t.Errorf("模板应被更新，得到 title=%q bodyA=%q", got.TitleTemplate, got.BodyTemplates[0])
	}
	// 关键有：词库、图片、变量不能被动。
	if strings.Join(got.Keywords, ",") != "关键词甲,关键词乙" {
		t.Errorf("SaveTemplates 不应动关键词，得到 %v", got.Keywords)
	}
	if strings.Join(got.Images, ",") != "图一" {
		t.Errorf("SaveTemplates 不应动图片库，得到 %v", got.Images)
	}
	if strings.Join(got.Vars[0], ",") != "变量甲" {
		t.Errorf("SaveTemplates 不应动变量库，得到 %v", got.Vars[0])
	}
}

func TestLoadSkipsBlankLinesAndTrimsCRLF(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, contentgen.Library{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, keywordsFile), []byte("甲\r\n\r\n  \r\n乙\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	lib, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}
	if len(lib.Keywords) != 2 || lib.Keywords[0] != "甲" || lib.Keywords[1] != "乙" {
		t.Errorf("应丢掉空行并去掉 CR，得到 %q", lib.Keywords)
	}
}

func TestLoadReadsArticleLibrary(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err != nil { // 先建出目录骨架
		t.Fatal(err)
	}

	files := map[string]string{
		"甲篇.txt": "甲的正文\n",
		"乙篇.md":  "乙的正文",
		"封面.png": "不该被读进来",
		"没扩展名":   "也不该被读进来",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, articlesDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	lib, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}
	if len(lib.Articles) != 2 {
		t.Fatalf("只应读进 .txt/.md 两篇，实际 %d 篇: %+v", len(lib.Articles), lib.Articles)
	}

	got := map[string]string{}
	for _, a := range lib.Articles {
		got[a.Name] = a.Body
	}
	if got["甲篇"] != "甲的正文" {
		t.Errorf("甲篇正文 = %q，期望 %q（结尾换行应被去掉）", got["甲篇"], "甲的正文")
	}
	if got["乙篇"] != "乙的正文" {
		t.Errorf("乙篇正文 = %q", got["乙篇"])
	}
}

func TestSaveLoadImageBank(t *testing.T) {
	dir := t.TempDir()
	want := []string{"https://example.com/a.png", "![图](https://example.com/b.png)"}

	if err := Save(dir, contentgen.Library{Images: want}); err != nil {
		t.Fatalf("Save 返回错误: %v", err)
	}
	lib, err := Load(dir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}
	if strings.Join(lib.Images, "|") != strings.Join(want, "|") {
		t.Errorf("图片库 = %v，期望 %v", lib.Images, want)
	}
}

// 文章库是用户往目录里丢文件，Save 不该把它清掉。
func TestSaveDoesNotTouchArticleLibrary(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, articlesDir, "手放进去的.txt")
	if err := os.WriteFile(path, []byte("内容"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Save(dir, contentgen.Library{}); err != nil {
		t.Fatalf("Save 返回错误: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Save 不应动文章库里的文件: %v", err)
	}
}

func TestExportDraftsWritesMarkdown(t *testing.T) {
	dir := t.TempDir()
	drafts := []contentgen.Draft{
		{Title: "第一篇", Body: "正文一"},
		{Title: "第二篇", Body: "正文二"},
	}

	n, err := ExportDrafts(dir, drafts)
	if err != nil {
		t.Fatalf("ExportDrafts 返回错误: %v", err)
	}
	if n != 2 {
		t.Errorf("应导出 2 篇，实际 %d 篇", n)
	}

	b, err := os.ReadFile(filepath.Join(dir, "第一篇.md"))
	if err != nil {
		t.Fatalf("读取导出文件失败: %v", err)
	}
	if got, want := string(b), "# 第一篇\n\n正文一\n"; got != want {
		t.Errorf("导出内容 = %q，期望 %q", got, want)
	}
}

func TestExportDraftsSanitizesFileName(t *testing.T) {
	dir := t.TempDir()
	drafts := []contentgen.Draft{{Title: `a/b\c:d*e?f"g<h>i|j`, Body: "正文"}}

	if _, err := ExportDrafts(dir, drafts); err != nil {
		t.Fatalf("ExportDrafts 返回错误: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("期望 1 个文件，实际 %d 个", len(entries))
	}
	name := entries[0].Name()
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		t.Errorf("文件名仍含非法字符: %q", name)
	}
	if got, want := name, "a_b_c_d_e_f_g_h_i_j.md"; got != want {
		t.Errorf("文件名 = %q，期望 %q", got, want)
	}
}

func TestExportDraftsDedupesNames(t *testing.T) {
	dir := t.TempDir()
	drafts := []contentgen.Draft{
		{Title: "同名", Body: "一"},
		{Title: "同名", Body: "二"},
		{Title: "同名", Body: "三"},
	}

	if _, err := ExportDrafts(dir, drafts); err != nil {
		t.Fatalf("ExportDrafts 返回错误: %v", err)
	}
	for _, name := range []string{"同名.md", "同名-2.md", "同名-3.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("期望存在 %s: %v", name, err)
		}
	}
}

func TestExportDraftsFallbackNameForUnusableTitle(t *testing.T) {
	dir := t.TempDir()
	drafts := []contentgen.Draft{
		{Title: "...", Body: "一"},
		{Title: "CON", Body: "二"},
	}

	if _, err := ExportDrafts(dir, drafts); err != nil {
		t.Fatalf("ExportDrafts 返回错误: %v", err)
	}
	for _, name := range []string{"草稿-1.md", "草稿-2.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("期望存在兜底文件名 %s: %v", name, err)
		}
	}
}

func TestExportDraftsTruncatesLongTitle(t *testing.T) {
	dir := t.TempDir()
	long := strings.Repeat("标", 300)

	if _, err := ExportDrafts(dir, []contentgen.Draft{{Title: long, Body: "正文"}}); err != nil {
		t.Fatalf("ExportDrafts 返回错误: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	name := strings.TrimSuffix(entries[0].Name(), ".md")
	if len([]rune(name)) != maxFileNameRunes {
		t.Errorf("文件名应截断到 %d 个字，实际 %d 个", maxFileNameRunes, len([]rune(name)))
	}
}

func TestExportDraftsRejectsEmpty(t *testing.T) {
	if _, err := ExportDrafts(t.TempDir(), nil); err == nil {
		t.Error("没有草稿时应返回错误")
	}
}

// TestEndToEndSaveGenerateExport 走一遍界面上的真实流程：
// 存素材 → 读回来生成 → 导出成 .md，确认三个环节接得上。
func TestEndToEndSaveGenerateExport(t *testing.T) {
	contentDir := filepath.Join(t.TempDir(), "content")
	exportDir := filepath.Join(t.TempDir(), "out")

	err := Save(contentDir, contentgen.Library{
		TitleTemplate: "{关键词}测评",
		BodyTemplates: []string{"{关键词}的要点：{变量1}。\n记录于 {日期1}。"},
		Keywords:      []string{"绿茶", "红茶"},
		Vars:          [][]string{{"回甘明显", "汤色透亮"}},
	})
	if err != nil {
		t.Fatalf("Save 返回错误: %v", err)
	}

	lib, err := Load(contentDir)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	now := time.Date(2026, 8, 25, 17, 42, 0, 0, time.UTC)
	drafts, err := contentgen.Generate(lib, contentgen.Options{Count: 2}, rand.New(rand.NewSource(1)), now)
	if err != nil {
		t.Fatalf("Generate 返回错误: %v", err)
	}
	if len(drafts) != 2 {
		t.Fatalf("期望 2 篇，实际 %d 篇", len(drafts))
	}
	if drafts[0].Title != "绿茶测评" || drafts[1].Title != "红茶测评" {
		t.Errorf("标题未按关键词轮转: %q / %q", drafts[0].Title, drafts[1].Title)
	}
	for _, d := range drafts {
		if strings.Contains(d.Body, "{") {
			t.Errorf("正文里还有没替换的占位符: %q", d.Body)
		}
		if !strings.Contains(d.Body, "2026-08-25") {
			t.Errorf("正文里没有注入日期: %q", d.Body)
		}
	}

	n, err := ExportDrafts(exportDir, drafts)
	if err != nil {
		t.Fatalf("ExportDrafts 返回错误: %v", err)
	}
	if n != 2 {
		t.Fatalf("期望导出 2 篇，实际 %d 篇", n)
	}

	b, err := os.ReadFile(filepath.Join(exportDir, "绿茶测评.md"))
	if err != nil {
		t.Fatalf("读取导出文件失败: %v", err)
	}
	if got := string(b); !strings.HasPrefix(got, "# 绿茶测评\n\n") || !strings.Contains(got, "绿茶的要点：") {
		t.Errorf("导出内容不符合预期: %q", got)
	}
}

func TestDefaultDirEndsWithContent(t *testing.T) {
	dir, err := DefaultDir()
	if err != nil {
		t.Skipf("当前环境拿不到用户配置目录: %v", err)
	}
	if got, want := filepath.Base(dir), "content"; got != want {
		t.Errorf("目录末级 = %q，期望 %q", got, want)
	}
	if got, want := filepath.Base(filepath.Dir(dir)), appName; got != want {
		t.Errorf("上级目录 = %q，期望 %q", got, want)
	}
}
