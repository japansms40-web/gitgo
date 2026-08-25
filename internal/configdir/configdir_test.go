package configdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildSample 建出一份接近真实「配置」目录的样例，返回其绝对路径。
func buildSample(t *testing.T) string {
	t.Helper()
	base := filepath.Join(t.TempDir(), "配置")
	dirs := []string{"变量", "文章", "文件库"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"关键词.txt":      "谷歌账号购买\n谷歌邮箱批发\n",
		"关键词1.txt":     "第二组关键词\n",
		"换号特征.txt":     "当日投稿数量已到达上限\n账号未登录\n",
		"变量/变量1.txt":   "娱乐新闻一\n娱乐新闻二\n",
		"文章/诗词abc.txt": "《黄鹤楼》\n故人西辞黄鹤楼\n",
		".DS_Store":    "junk",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(base, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return base
}

func findNode(nodes []Node, name string) (Node, bool) {
	for _, n := range nodes {
		if n.Name == name {
			return n, true
		}
	}
	return Node{}, false
}

func TestTreeListsDirsBeforeFilesAndSkipsHidden(t *testing.T) {
	base := buildSample(t)

	nodes, err := Tree(base)
	if err != nil {
		t.Fatalf("Tree 返回错误: %v", err)
	}

	// 隐藏文件不出现。
	if _, ok := findNode(nodes, ".DS_Store"); ok {
		t.Error(".DS_Store 这类隐藏文件不应出现在树里")
	}

	// 目录排在文件前面。
	firstFileSeen := false
	for _, n := range nodes {
		if !n.IsDir {
			firstFileSeen = true
		} else if firstFileSeen {
			t.Errorf("目录 %s 应排在所有文件之前", n.Name)
		}
	}

	// 子目录被递归展开。
	varsDir, ok := findNode(nodes, "变量")
	if !ok || !varsDir.IsDir {
		t.Fatal("应包含「变量」目录节点")
	}
	if _, ok := findNode(varsDir.Children, "变量1.txt"); !ok {
		t.Error("「变量」目录下应递归列出 变量1.txt")
	}
	// 子节点 Path 用 / 分隔且相对根。
	child, _ := findNode(varsDir.Children, "变量1.txt")
	if child.Path != "变量/变量1.txt" {
		t.Errorf("子节点 Path 应为 变量/变量1.txt，实际 %q", child.Path)
	}
}

func TestReadFileReturnsContent(t *testing.T) {
	base := buildSample(t)

	preview, err := ReadFile(base, "换号特征.txt")
	if err != nil {
		t.Fatalf("ReadFile 返回错误: %v", err)
	}
	if !strings.Contains(preview.Content, "当日投稿数量已到达上限") {
		t.Errorf("应读到换号特征内容，实际 %q", preview.Content)
	}
	if preview.Truncated {
		t.Error("小文件不应被标记为截断")
	}
}

func TestReadFileRejectsTraversal(t *testing.T) {
	base := buildSample(t)
	// 在配置目录外放一个秘密文件。
	secret := filepath.Join(filepath.Dir(base), "secret.txt")
	if err := os.WriteFile(secret, []byte("绝密"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadFile(base, "../secret.txt"); err == nil {
		t.Error("用 ../ 读取配置目录外的文件应被拒绝")
	}
}

func TestReadFileRejectsDir(t *testing.T) {
	base := buildSample(t)
	if _, err := ReadFile(base, "变量"); err == nil {
		t.Error("对目录调用 ReadFile 应报错")
	}
}

func TestReadFileTruncatesLargeFile(t *testing.T) {
	base := buildSample(t)
	big := strings.Repeat("这是一行很长的中文内容用于撑大文件。\n", 40000)
	if err := os.WriteFile(filepath.Join(base, "大文件.txt"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}

	preview, err := ReadFile(base, "大文件.txt")
	if err != nil {
		t.Fatalf("ReadFile 返回错误: %v", err)
	}
	if !preview.Truncated {
		t.Error("超过上限的文件应被标记为截断")
	}
	if len(preview.Content) > maxPreviewBytes {
		t.Errorf("截断后内容字节数 %d 不应超过上限 %d", len(preview.Content), maxPreviewBytes)
	}
}

func TestWriteFileUpdatesContent(t *testing.T) {
	base := buildSample(t)

	if err := WriteFile(base, "关键词.txt", "新关键词一\n新关键词二\n"); err != nil {
		t.Fatalf("WriteFile 返回错误: %v", err)
	}
	preview, err := ReadFile(base, "关键词.txt")
	if err != nil {
		t.Fatalf("ReadFile 返回错误: %v", err)
	}
	if preview.Content != "新关键词一\n新关键词二\n" {
		t.Errorf("写入后内容 = %q", preview.Content)
	}
}

func TestWriteFileNormalizesCRLF(t *testing.T) {
	base := buildSample(t)
	if err := WriteFile(base, "关键词.txt", "甲\r\n乙\r\n"); err != nil {
		t.Fatalf("WriteFile 返回错误: %v", err)
	}
	preview, _ := ReadFile(base, "关键词.txt")
	if strings.Contains(preview.Content, "\r") {
		t.Errorf("落盘应统一为 \\n，实际 %q", preview.Content)
	}
}

func TestWriteFileRejectsTraversal(t *testing.T) {
	base := buildSample(t)
	secret := filepath.Join(filepath.Dir(base), "secret.txt")
	if err := os.WriteFile(secret, []byte("原始"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(base, "../secret.txt", "被篡改"); err == nil {
		t.Error("用 ../ 写配置目录外的文件应被拒绝")
	}
	got, _ := os.ReadFile(secret)
	if string(got) != "原始" {
		t.Error("目录外文件不应被写入")
	}
}

func TestWriteFileRejectsMissingFile(t *testing.T) {
	base := buildSample(t)
	if err := WriteFile(base, "不存在.txt", "x"); err == nil {
		t.Error("写不存在的文件应报错（只允许改已有文件）")
	}
}

func TestWriteFileRejectsDir(t *testing.T) {
	base := buildSample(t)
	if err := WriteFile(base, "变量", "x"); err == nil {
		t.Error("对目录调用 WriteFile 应报错")
	}
}

func TestResolveWalksUpFromSubdir(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "docs", "git", "配置")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	// 从深层子目录出发，相对路径应能向上找到配置目录。
	sub := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)

	got, err := Resolve(filepath.FromSlash("docs/git/配置"))
	if err != nil {
		t.Fatalf("Resolve 返回错误: %v", err)
	}
	if got != target {
		t.Errorf("应解析到 %s，实际 %s", target, got)
	}
}

func TestResolveEmptyErrors(t *testing.T) {
	if _, err := Resolve("  "); err == nil {
		t.Error("空配置目录应报错")
	}
}
