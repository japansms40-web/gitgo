package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestListRepos_DecodeResponse 离线验证响应结构体：拿 01_GET_repos.txt 里抠出的
// 真实响应体喂给结构体，断言关键字段解得对。永远稳定、不依赖网络。
func TestListRepos_DecodeResponse(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "repos_response.json"))
	if err != nil {
		t.Fatalf("读取 testdata 失败: %v", err)
	}

	var resp ReposResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	route := resp.Payload.ReposFinderPageRoute
	if route.RepositoryCount != 1 {
		t.Errorf("RepositoryCount = %d, 期望 1", route.RepositoryCount)
	}
	if route.PageCount != 1 {
		t.Errorf("PageCount = %d, 期望 1", route.PageCount)
	}
	if len(route.Repositories) != 1 {
		t.Fatalf("Repositories 长度 = %d, 期望 1", len(route.Repositories))
	}

	repo := route.Repositories[0]
	if repo.Name != "abyuds" {
		t.Errorf("Name = %q, 期望 %q", repo.Name, "abyuds")
	}
	if repo.Type != "Public" {
		t.Errorf("Type = %q, 期望 %q", repo.Type, "Public")
	}
	if repo.Owner != "VKMW57by18" {
		t.Errorf("Owner = %q, 期望 %q", repo.Owner, "VKMW57by18")
	}
	if !repo.LastUpdated.HasBeenPushedTo {
		t.Errorf("LastUpdated.HasBeenPushedTo = false, 期望 true")
	}
	if repo.LastUpdated.Timestamp != "2026-08-23T16:10:15.134Z" {
		t.Errorf("LastUpdated.Timestamp = %q", repo.LastUpdated.Timestamp)
	}
}
