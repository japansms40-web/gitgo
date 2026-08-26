package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateRepo_DecodeResponse 离线验证建仓库响应结构体：拿真实响应体
// （TestCreateRepo_Live 实测抓到）喂给结构体，断言 redirect 解得对。
func TestCreateRepo_DecodeResponse(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "create_repo_response.json"))
	if err != nil {
		t.Fatalf("读取 testdata 失败: %v", err)
	}

	var resp CreateRepoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if !strings.HasPrefix(resp.Data.Redirect, "/aA6QyMrh1k78/claude-test-repo-") {
		t.Errorf("Data.Redirect = %q", resp.Data.Redirect)
	}
}

// TestBuildCreateRepoBody 离线验证建仓库的 JSON body：不发网络，断言结构与固定字段对齐抓包
// docs 参考 scratch（POST /repositories）。
func TestBuildCreateRepoBody(t *testing.T) {
	p := CreateRepoParams{
		Owner:       "octocat",
		Name:        "hello-world",
		Visibility:  "public",
		Description: "demo",
	}

	raw, err := buildCreateRepoBody(p)
	if err != nil {
		t.Fatalf("buildCreateRepoBody 失败: %v", err)
	}

	// 解回通用结构，逐项核对。
	var got struct {
		Owner              string `json:"owner"`
		TemplateRepoID     string `json:"template_repository_id"`
		IncludeAllBranches string `json:"include_all_branches"`
		Repository         struct {
			Name              string `json:"name"`
			Visibility        string `json:"visibility"`
			Description       string `json:"description"`
			AutoInit          string `json:"auto_init"`
			LicenseTemplate   string `json:"license_template"`
			GitignoreTemplate string `json:"gitignore_template"`
		} `json:"repository"`
		Metrics map[string]any `json:"metrics"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("生成的 body 不是合法 JSON: %v", err)
	}

	if got.Owner != "octocat" {
		t.Errorf("owner = %q", got.Owner)
	}
	if got.Repository.Name != "hello-world" {
		t.Errorf("repository.name = %q", got.Repository.Name)
	}
	if got.Repository.Visibility != "public" {
		t.Errorf("repository.visibility = %q", got.Repository.Visibility)
	}
	if got.Repository.Description != "demo" {
		t.Errorf("repository.description = %q", got.Repository.Description)
	}
	// 固定字段（复刻抓包）。
	if got.Repository.AutoInit != "0" {
		t.Errorf("auto_init = %q, 期望 0", got.Repository.AutoInit)
	}
	if got.IncludeAllBranches != "0" {
		t.Errorf("include_all_branches = %q, 期望 0", got.IncludeAllBranches)
	}
	if got.Metrics == nil {
		t.Errorf("metrics 缺失")
	}
	if v, ok := got.Metrics["submitted_using_v2"]; !ok || v != true {
		t.Errorf("metrics.submitted_using_v2 = %v, 期望 true", v)
	}
}
