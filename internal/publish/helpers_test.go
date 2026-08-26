package publish

import "testing"

func TestCommitSHAFromQuorumPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/aA6QyMrh1k78/repo/check_commit_quorum/4cb211608e4e2849d92118952c8116790052b8e9", "4cb211608e4e2849d92118952c8116790052b8e9"},
		{"/o/r/check_commit_quorum/abc123", "abc123"},
		{"", ""},
		{"no-slash", "no-slash"},
		{"/trailing/slash/", ""},
	}
	for _, c := range cases {
		if got := commitSHAFromQuorumPath(c.in); got != c.want {
			t.Errorf("commitSHAFromQuorumPath(%q) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeRepoName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"hello-world", "hello-world"},
		{"谷歌邮箱 批发", "谷歌邮箱-批发"},      // 空格→连字符（GitHub 允许非 ASCII 会被转码，但空格必须换）
		{"a/b\\c:d", "a-b-c-d"},     // 非法路径/分隔字符→连字符
		{"  trim--me  ", "trim-me"}, // 去首尾、折叠多个连字符
		{"", "repo"},                // 空→兜底
		{"...", "repo"},             // 全是点→兜底
	}
	for _, c := range cases {
		if got := sanitizeRepoName(c.in); got != c.want {
			t.Errorf("sanitizeRepoName(%q) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	// 文件名：去换行、去路径分隔、非空、保证 .md 结尾。
	if got := sanitizeFilename("标题\n换行"); got != "标题-换行.md" {
		t.Errorf("换行处理错: %q", got)
	}
	if got := sanitizeFilename("a/b"); got != "a-b.md" {
		t.Errorf("斜杠处理错: %q", got)
	}
	if got := sanitizeFilename(""); got != "index.md" {
		t.Errorf("空兜底错: %q", got)
	}
	if got := sanitizeFilename("已带后缀.md"); got != "已带后缀.md" {
		t.Errorf("已有 .md 不该重复: %q", got)
	}
}
