// Package configdir 负责读取「配置目录」——用户放关键词、变量、文章、文件库、
// 换号特征等 txt 的那份目录（默认 docs/git/配置）。只做只读浏览与预览，
// 供界面上的「文件库」标签页把整棵目录树展示出来、点开某个文件看内容。
package configdir

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// maxPreviewBytes 是单个文件预览返回的最大字节数；超过就截断，避免把变量库这类
// 上万行的大文件整份塞给前端。
const maxPreviewBytes = 256 * 1024

// Node 是配置目录里的一个文件或子目录节点，Path 用 / 分隔且相对配置目录根。
type Node struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Children []Node `json:"children,omitempty"`
}

// FilePreview 是一个文件的预览内容。Truncated 为真表示内容因过大被截断。
type FilePreview struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated"`
}

// Resolve 把前端传来的配置目录解析成一个真实存在的绝对路径。
// 绝对路径直接校验；相对路径先按当前工作目录找，找不到再沿父目录逐级向上寻找，
// 兼容 wails dev 从项目根或子目录启动的情况。
func Resolve(configPath string) (string, error) {
	if strings.TrimSpace(configPath) == "" {
		return "", fmt.Errorf("配置目录为空")
	}

	if filepath.IsAbs(configPath) {
		if isDir(configPath) {
			return filepath.Clean(configPath), nil
		}
		return "", fmt.Errorf("配置目录不存在：%s", configPath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; {
		candidate := filepath.Join(dir, configPath)
		if isDir(candidate) {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("找不到配置目录：%s", configPath)
}

// Tree 返回配置目录下的文件树（根目录本身不作为节点，直接返回它的子节点）。
// 同级里目录排在文件前面，各自按名称排序，界面渲染顺序稳定。
func Tree(configPath string) ([]Node, error) {
	base, err := Resolve(configPath)
	if err != nil {
		return nil, err
	}
	return readChildren(base, "")
}

// ReadFile 读取配置目录下 relPath 指向的文件内容用于预览。
// relPath 必须落在配置目录内，防止用 .. 读到目录外的文件。
func ReadFile(configPath, relPath string) (FilePreview, error) {
	return readPreview(configPath, relPath, false)
}

// ReadFileTail 与 ReadFile 相同，但文件过大时保留尾部而非头部。供「查看链接」这类
// 边发边 append、最新内容在末尾的大文件用——头部截断会只显示最旧的链接。
func ReadFileTail(configPath, relPath string) (FilePreview, error) {
	return readPreview(configPath, relPath, true)
}

// ResolveFile 把配置目录下的相对路径解析成校验过的绝对路径，供在文件管理器里定位。
func ResolveFile(configPath, relPath string) (string, error) {
	base, err := Resolve(configPath)
	if err != nil {
		return "", err
	}
	return safeJoin(base, relPath)
}

func readPreview(configPath, relPath string, tail bool) (FilePreview, error) {
	base, err := Resolve(configPath)
	if err != nil {
		return FilePreview{}, err
	}
	full, err := safeJoin(base, relPath)
	if err != nil {
		return FilePreview{}, err
	}

	info, err := os.Stat(full)
	if err != nil {
		return FilePreview{}, err
	}
	if info.IsDir() {
		return FilePreview{}, fmt.Errorf("%s 是目录，不能预览", relPath)
	}

	b, err := os.ReadFile(full)
	if err != nil {
		return FilePreview{}, err
	}

	truncated := false
	if len(b) > maxPreviewBytes {
		truncated = true
		if tail {
			b = b[len(b)-maxPreviewBytes:]
			// 从首个换行之后开始：丢掉开头那半行（顺带去掉被切断的多字节字符），
			// 保证预览从一整行的行首起。
			if i := bytes.IndexByte(b, '\n'); i >= 0 {
				b = b[i+1:]
			}
		} else {
			b = b[:maxPreviewBytes]
			// 截断可能切断多字节字符，去掉尾部残缺的 UTF-8 字节。
			for len(b) > 0 && !utf8.Valid(b) {
				b = b[:len(b)-1]
			}
		}
	}

	content := strings.ReplaceAll(string(b), "\r\n", "\n")
	rel, _ := filepath.Rel(base, full)
	return FilePreview{Path: filepath.ToSlash(rel), Content: content, Truncated: truncated}, nil
}

// WriteFile 把 content 写回配置目录下 relPath 指向的文件（供「文件库」里直接编辑保存）。
// 只允许写已存在的文件，且必须落在配置目录内；不新建文件、不覆盖目录。
func WriteFile(configPath, relPath, content string) error {
	base, err := Resolve(configPath)
	if err != nil {
		return err
	}
	full, err := safeJoin(base, relPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(full)
	if err != nil {
		return err // 文件不存在等错误如实返回
	}
	if info.IsDir() {
		return fmt.Errorf("%s 是目录，不能写入", relPath)
	}

	// 统一成 \n 落盘，跨平台一致。
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	return os.WriteFile(full, []byte(normalized), 0o644)
}

// safeJoin 把相对路径拼到配置目录下，并确认结果仍在目录内，挡住 ../ 穿越。
func safeJoin(base, relPath string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if clean == "." || clean == string(filepath.Separator) {
		return "", fmt.Errorf("未指定文件")
	}
	full := filepath.Join(base, clean)
	rel, err := filepath.Rel(base, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("非法路径：%s", relPath)
	}
	return full, nil
}

// readChildren 递归读取 dir（base 下相对路径为 relDir）的直接子节点。
func readChildren(base, relDir string) ([]Node, error) {
	entries, err := os.ReadDir(filepath.Join(base, relDir))
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue // 跳过 .DS_Store 之类的隐藏文件
		}
		relPath := name
		if relDir != "" {
			relPath = filepath.ToSlash(filepath.Join(relDir, name))
		}
		node := Node{Name: name, Path: relPath, IsDir: entry.IsDir()}
		if entry.IsDir() {
			children, err := readChildren(base, filepath.FromSlash(relPath))
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		nodes = append(nodes, node)
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir // 目录排在文件前
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
