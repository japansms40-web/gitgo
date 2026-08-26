package github

import "testing"

// TestOwnerFromCookie 从会话 Cookie 串里取 dotcom_user 作为 owner。
func TestOwnerFromCookie(t *testing.T) {
	cases := []struct {
		name   string
		cookie string
		want   string
	}{
		{"正常", "_gh_sess=abc; user_session=xyz; dotcom_user=octocat", "octocat"},
		{"结尾无分号", "foo=1; dotcom_user=aA6QyMrh1k78", "aA6QyMrh1k78"},
		{"带连字符", "dotcom_user=chenweilong1022-alt; foo=1", "chenweilong1022-alt"},
		{"缺失返回空", "_gh_sess=abc; user_session=xyz", ""},
		{"空串返回空", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := OwnerFromCookie(tc.cookie); got != tc.want {
				t.Errorf("OwnerFromCookie(%q) = %q, 期望 %q", tc.cookie, got, tc.want)
			}
		})
	}
}
