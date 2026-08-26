package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestCreateRepoWithFile_Live 端到端跑编排：真建一个新仓库并写入首个文件——【写操作】，双重开关守卫
// （GH_TEST_COOKIE + GH_TEST_ALLOW_MUTATE=1）。同时验证「空仓库首次提交」（BaseCommit 留空）可行。
//
// 跑法：
//
//	GH_TEST_COOKIE='...' GH_TEST_ALLOW_MUTATE=1 go test -run TestCreateRepoWithFile_Live ./internal/github/
func TestCreateRepoWithFile_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需真实网络与有效 Cookie，-short 模式跳过")
	}
	if os.Getenv("GH_TEST_ALLOW_MUTATE") != "1" {
		t.Skip("写操作测试默认跳过：设 GH_TEST_ALLOW_MUTATE=1 才真建仓库+文件")
	}

	cookie := testCookie(t)
	owner := os.Getenv("GH_TEST_OWNER")
	if owner == "" {
		owner = OwnerFromCookie(cookie)
	}
	if owner == "" {
		t.Skip("无法确定 owner：设 GH_TEST_OWNER 或让 Cookie 带 dotcom_user")
	}

	var opts []Option
	if p := testProxy(); p != "" {
		t.Logf("走代理: %s", p)
		opts = append(opts, WithProxy(p))
	}
	c, err := New(cookie, opts...)
	if err != nil {
		t.Fatalf("构造客户端失败: %v", err)
	}

	repoName := fmt.Sprintf("claude-rwf-%d", time.Now().Unix())
	res, err := c.CreateRepoWithFile(context.Background(), CreateRepoWithFileParams{
		Owner:         owner,
		RepoName:      repoName,
		Visibility:    "public",
		Filename:      "README.md",
		Content:       "# " + repoName + "\n由 CreateRepoWithFile 编排一步建仓库+首文件。",
		CommitMessage: "init: 首个文件",
	})
	if err != nil {
		t.Fatalf("CreateRepoWithFile 失败: %v", err)
	}

	if !strings.Contains(res.Repo.Data.Redirect, owner+"/"+repoName) {
		t.Errorf("Repo.Redirect = %q, 期望含 %q", res.Repo.Data.Redirect, owner+"/"+repoName)
	}
	if !strings.Contains(res.File.Data.Redirect, owner+"/"+repoName) {
		t.Errorf("File.Redirect = %q, 期望含 %q", res.File.Data.Redirect, owner+"/"+repoName)
	}
	t.Logf("一步建成：仓库 %s，文件跳转 %s", res.Repo.Data.Redirect, res.File.Data.Redirect)
}
