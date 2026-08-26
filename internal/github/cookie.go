package github

import "regexp"

// dotcomUserPattern 匹配会话 Cookie 里的 dotcom_user=<login>。
var dotcomUserPattern = regexp.MustCompile(`dotcom_user=([^;]+)`)

// OwnerFromCookie 从整串会话 Cookie 里取 dotcom_user 作为账号 owner（登录名）；取不到返回空。
// 建仓库/发文件需要 owner，而空账号 ListRepos 拿不到 owner，只能从 Cookie 推断。
func OwnerFromCookie(cookie string) string {
	m := dotcomUserPattern.FindStringSubmatch(cookie)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}
