package article

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Article 是队列中的一篇待发布文章。
type Article struct {
	Title     string // 文件名去扩展名，作为展示标题
	LocalPath string
	RepoPath  string // 仓库内路径，如 posts/hello.md
	Content   []byte
}

// ScanPaths 接收文件或文件夹路径列表，递归收集 .md/.txt，
// 生成发布队列。.txt 的仓库路径改为 .md。dir 为仓库内目标目录（可空）。
func ScanPaths(paths []string, dir string) ([]Article, error) {
	var out []Article
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err = filepath.WalkDir(p, func(fp string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if a, ok := buildArticle(fp, dir); ok {
					out = append(out, a)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if a, ok := buildArticle(p, dir); ok {
				out = append(out, a)
			}
		}
	}
	return out, nil
}

func buildArticle(localPath, dir string) (Article, bool) {
	ext := strings.ToLower(filepath.Ext(localPath))
	if ext != ".md" && ext != ".txt" {
		return Article{}, false
	}
	content, err := os.ReadFile(localPath)
	if err != nil {
		return Article{}, false
	}
	base := filepath.Base(localPath)
	title := strings.TrimSuffix(base, filepath.Ext(base))
	repoName := title + ".md" // 统一为 .md
	repoPath := repoName
	if dir != "" {
		repoPath = path.Join(dir, repoName) // 用 path 保证仓库内用 / 分隔
	}
	return Article{
		Title:     title,
		LocalPath: localPath,
		RepoPath:  repoPath,
		Content:   content,
	}, true
}
