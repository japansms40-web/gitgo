package publish

import (
	"strings"
	"unicode"
)

// commitSHAFromQuorumPath 从 CreateFile 响应的 commitQuorumPollPath
// （形如 /owner/repo/check_commit_quorum/<sha>）里抽出新 commit sha，用作下一篇的父提交。
func commitSHAFromQuorumPath(path string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return path
	}
	return path[i+1:]
}

// sanitizeRepoName 把任意标题净化成 GitHub 合法仓库名：保留字母/数字/‘-’‘_’‘.’，
// 其余（空格、路径分隔、符号）转 ‘-’，折叠连续 ‘-’，去首尾 ‘-’‘.’，空则兜底 "repo"。
func sanitizeRepoName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := collapseDashes(b.String())
	out = strings.Trim(out, "-.")
	if out == "" {
		return "repo"
	}
	return out
}

// sanitizeFilename 把标题净化成安全的 .md 文件名：换行/路径分隔转 ‘-’，去首尾空白，
// 空则兜底 "index"，保证 .md 结尾。
func sanitizeFilename(s string) string {
	repl := strings.NewReplacer("\n", "-", "\r", "-", "/", "-", "\\", "-")
	s = strings.TrimSpace(repl.Replace(s))
	if s == "" {
		s = "index"
	}
	if !strings.HasSuffix(strings.ToLower(s), ".md") {
		s += ".md"
	}
	return s
}

// collapseDashes 把连续的 ‘-’ 折叠成一个。
func collapseDashes(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if r == '-' {
			if !prevDash {
				b.WriteRune(r)
			}
			prevDash = true
			continue
		}
		b.WriteRune(r)
		prevDash = false
	}
	return b.String()
}
