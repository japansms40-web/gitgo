package article

import (
	"os"
	"path/filepath"
	"strings"
)

// Article 是队列中的一篇待发布内容。
type Article struct {
	Title     string // 文件名去扩展名，作为展示标题
	LocalPath string
	Content   []byte
}

// ScanPaths 接收文件或文件夹路径列表，递归收集 .md/.txt，生成发布队列。
func ScanPaths(paths []string) ([]Article, error) {
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
				if a, ok := buildArticle(fp); ok {
					out = append(out, a)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if a, ok := buildArticle(p); ok {
				out = append(out, a)
			}
		}
	}
	return out, nil
}

func buildArticle(localPath string) (Article, bool) {
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
	return Article{
		Title:     title,
		LocalPath: localPath,
		Content:   content,
	}, true
}
