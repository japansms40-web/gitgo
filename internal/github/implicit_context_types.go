package github

// ImplicitContext 是 GitHub 网页版 Copilot 聊天「隐式上下文」接口的响应：
// GET /github-copilot/chat/implicit-context/{owner}/{repo}/{urlencode(页面路径)}。
// 字段对着 docs/flows_1_requests/11_GET_github-copilot-chat-implicit-context-GhHnCQtQMat.txt。
type ImplicitContext struct {
	// Type 上下文类型，如 "file"。
	Type string `json:"type"`
	// URL 当前页面对应的仓库内路径，如 "/owner/repo/new/main"。
	URL string `json:"url"`
	// Path 具体文件路径；抓包为 null，用指针区分「无」与空串。
	Path *string `json:"path"`
	// RepoID 仓库数值 ID。
	RepoID int64 `json:"repoID"`
	// RepoOwner 仓库属主登录名。
	RepoOwner string `json:"repoOwner"`
	// RepoName 仓库名。
	RepoName string `json:"repoName"`
	// Ref 分支/标签名，如 "main"。
	Ref string `json:"ref"`
	// CommitOID 当前 ref 指向的 commit SHA。
	CommitOID string `json:"commitOID"`
}
