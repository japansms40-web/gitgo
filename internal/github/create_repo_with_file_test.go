package github

import (
	"context"
	"strings"
	"testing"
)

// TestCreateRepoWithFile_Validation 离线验证参数校验：必填项缺失时在发任何网络请求【之前】就返回
// error（用 dummy cookie 构造客户端，只走校验分支，不联网）。
func TestCreateRepoWithFile_Validation(t *testing.T) {
	c, err := New("dummy=1")
	if err != nil {
		t.Fatalf("构造客户端失败: %v", err)
	}

	cases := []struct {
		name string
		p    CreateRepoWithFileParams
		want string // error 里应含的关键字
	}{
		{"缺 owner", CreateRepoWithFileParams{RepoName: "r", Filename: "a.md"}, "owner"},
		{"缺 repo", CreateRepoWithFileParams{Owner: "o", Filename: "a.md"}, "name"},
		{"缺 filename", CreateRepoWithFileParams{Owner: "o", RepoName: "r"}, "filename"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.CreateRepoWithFile(context.Background(), tc.p)
			if err == nil {
				t.Fatalf("期望报错，实际 nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, 期望含 %q", err.Error(), tc.want)
			}
		})
	}
}
