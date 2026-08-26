// Package contentstore 负责把生成素材以 txt 文件的形式存取，
// 以及把生成好的草稿导出成 Markdown 文件。
//
// 素材目录布局（每个字段就是一个可以直接用文本编辑器改的 txt）：
//
//	content/
//	├── 标题模板.txt
//	├── 正文模板A.txt
//	├── 正文模板B.txt
//	├── 关键词.txt
//	└── 变量/
//	    └── 变量1.txt … 变量5.txt
package contentstore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitmd/internal/contentgen"
)

const appName = "gitmd"

const (
	titleFile    = "标题模板.txt"
	keywordsFile = "关键词.txt"
	imagesFile   = "图片库.txt"
	urlsFile     = "外链库.txt"
	varsDir      = "变量"
	articlesDir  = "文章库"
)

// articleExts 是文章库里会被当作素材读进来的扩展名。
var articleExts = map[string]bool{".txt": true, ".md": true}

// bodyLabels 给正文模板编号，对应界面上的 A / B 两个框。
var bodyLabels = [contentgen.BodyTemplateCount]string{"A", "B"}

// maxFileNameRunes 限制导出文件名长度，避免超长标题撑爆文件系统限制。
const maxFileNameRunes = 80

// DefaultDir 返回本机默认的素材目录。
func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName, "content"), nil
}

// Load 读取素材目录下的模板与词库。目录或文件不存在时会先建出带示例内容的骨架，
// 这样第一次打开就能直接生成，也能看出该往哪些文件里填东西。
func Load(dir string) (contentgen.Library, error) {
	var lib contentgen.Library
	if err := ensureLayout(dir); err != nil {
		return lib, err
	}

	var err error
	if lib.TitleTemplate, err = readText(filepath.Join(dir, titleFile)); err != nil {
		return lib, err
	}

	lib.BodyTemplates = make([]string, contentgen.BodyTemplateCount)
	for i := range lib.BodyTemplates {
		if lib.BodyTemplates[i], err = readText(bodyPath(dir, i)); err != nil {
			return lib, err
		}
	}

	keywords, err := readText(filepath.Join(dir, keywordsFile))
	if err != nil {
		return lib, err
	}
	lib.Keywords = splitLines(keywords)

	images, err := readText(filepath.Join(dir, imagesFile))
	if err != nil {
		return lib, err
	}
	lib.Images = splitLines(images)

	urls, err := readText(filepath.Join(dir, urlsFile))
	if err != nil {
		return lib, err
	}
	lib.URLs = splitLines(urls)

	if lib.Articles, err = loadArticles(filepath.Join(dir, articlesDir)); err != nil {
		return lib, err
	}

	// 始终返回固定数量的词库槽位，前端可以直接按下标渲染 5 个输入框。
	lib.Vars = make([][]string, contentgen.VarBankCount)
	for i := range lib.Vars {
		text, err := readText(varPath(dir, i+1))
		if err != nil {
			return lib, err
		}
		lib.Vars[i] = splitLines(text)
	}
	return lib, nil
}

// Save 把模板与词库写回 txt 文件。
func Save(dir string, lib contentgen.Library) error {
	if err := ensureLayout(dir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, titleFile), []byte(lib.TitleTemplate), 0o644); err != nil {
		return err
	}
	for i := 0; i < contentgen.BodyTemplateCount; i++ {
		var body string
		if i < len(lib.BodyTemplates) {
			body = lib.BodyTemplates[i]
		}
		if err := os.WriteFile(bodyPath(dir, i), []byte(body), 0o644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(dir, keywordsFile), []byte(strings.Join(lib.Keywords, "\n")), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, imagesFile), []byte(strings.Join(lib.Images, "\n")), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, urlsFile), []byte(strings.Join(lib.URLs, "\n")), 0o644); err != nil {
		return err
	}
	// lib.Articles 不回写：文章库是用户往「文章库」目录里丢文件，只读不改。
	// 无论传进来几组，磁盘上始终维持 VarBankCount 个词库文件。
	for i := 0; i < contentgen.VarBankCount; i++ {
		var bank []string
		if i < len(lib.Vars) {
			bank = lib.Vars[i]
		}
		if err := os.WriteFile(varPath(dir, i+1), []byte(strings.Join(bank, "\n")), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// SaveTemplates 只写标题模板与正文模板 A/B，不动关键词/图片/变量/文章等数据文件。
// 界面上「变量设置」已并入「文件库」，词库改由用户在文件库里直接改文件，
// 所以生成前落盘时只需要把模板存回去，避免把文件库里的编辑覆盖掉。
func SaveTemplates(dir, title string, bodies []string) error {
	if err := ensureLayout(dir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, titleFile), []byte(title), 0o644); err != nil {
		return err
	}
	for i := 0; i < contentgen.BodyTemplateCount; i++ {
		var body string
		if i < len(bodies) {
			body = bodies[i]
		}
		if err := os.WriteFile(bodyPath(dir, i), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ExportDrafts 把草稿逐篇写成 <dir>/<标题>.md，返回成功写出的篇数。
func ExportDrafts(dir string, drafts []contentgen.Draft) (int, error) {
	if len(drafts) == 0 {
		return 0, errors.New("没有可导出的草稿")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}

	used := make(map[string]bool, len(drafts))
	written := 0
	for i, draft := range drafts {
		name := uniqueName(sanitizeFileName(draft.Title, i), used)
		content := "# " + draft.Title + "\n\n" + draft.Body + "\n"
		if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0o644); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

// ensureLayout 建出素材目录与其中缺失的文件，缺的文件填入示例内容；
// 已存在的文件不会被动。
func ensureLayout(dir string) error {
	for _, sub := range []string{varsDir, articlesDir} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return err
		}
	}

	seed := map[string]string{
		filepath.Join(dir, titleFile):    defaultTitleTemplate,
		filepath.Join(dir, keywordsFile): defaultKeywords,
		filepath.Join(dir, imagesFile):   defaultImages,
		filepath.Join(dir, urlsFile):     defaultURLs,
		bodyPath(dir, 0):                 defaultBodyTemplateA,
		bodyPath(dir, 1):                 defaultBodyTemplateB,
	}
	for i := 1; i <= contentgen.VarBankCount; i++ {
		var content string
		if i <= len(defaultVarContents) {
			content = defaultVarContents[i-1]
		}
		seed[varPath(dir, i)] = content
	}

	for path, content := range seed {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return err
		}
		_, writeErr := f.WriteString(content)
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

// loadArticles 读「文章库」目录下的 .txt/.md，文件名去扩展名作为 {文章名}。
// 目录为空是正常情况，返回空切片。
func loadArticles(dir string) ([]contentgen.Article, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var articles []contentgen.Article
	for _, entry := range entries {
		if entry.IsDir() || !articleExts[strings.ToLower(filepath.Ext(entry.Name()))] {
			continue
		}
		body, err := readText(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		articles = append(articles, contentgen.Article{
			Name: strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Body: strings.TrimRight(body, "\n"),
		})
	}
	return articles, nil
}

func bodyPath(dir string, index int) string {
	return filepath.Join(dir, fmt.Sprintf("正文模板%s.txt", bodyLabels[index]))
}

func varPath(dir string, n int) string {
	return filepath.Join(dir, varsDir, fmt.Sprintf("变量%d.txt", n))
}

// readText 读取文件内容；文件不存在按空内容处理。
func readText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n"), nil
}

// splitLines 把词库文本切成一行一条，丢掉空行。
func splitLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// windowsReserved 是 Windows 上不能作为文件名的保留字（跨平台一并回避）。
var windowsReserved = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// sanitizeFileName 把标题变成各平台都能落地的文件名；index 用于兜底命名。
func sanitizeFileName(title string, index int) string {
	var b strings.Builder
	for _, r := range title {
		switch {
		case strings.ContainsRune(`/\:*?"<>|`, r), r < 0x20, r == 0x7f:
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	// Windows 会自动去掉结尾的点和空格，这里先去掉以免出现意料外的重名。
	name := strings.TrimRight(strings.TrimSpace(b.String()), ". ")

	if runes := []rune(name); len(runes) > maxFileNameRunes {
		name = strings.TrimRight(string(runes[:maxFileNameRunes]), ". ")
	}
	if name == "" || windowsReserved[strings.ToUpper(name)] {
		name = fmt.Sprintf("草稿-%d", index+1)
	}
	return name
}

// uniqueName 给重名的文件加 -2、-3 后缀。
func uniqueName(name string, used map[string]bool) string {
	candidate := name
	for i := 2; used[strings.ToLower(candidate)]; i++ {
		candidate = fmt.Sprintf("%s-%d", name, i)
	}
	used[strings.ToLower(candidate)] = true
	return candidate
}
