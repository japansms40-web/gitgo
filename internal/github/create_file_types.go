package github

// CreateFileParams 是「创建文件并提交」的入参。
type CreateFileParams struct {
	// Owner 仓库属主登录名。
	Owner string
	// Repo 仓库名。
	Repo string
	// Branch 目标分支，如 "main"。同时作为 URL 末段与表单 target_branch。
	Branch string
	// Filename 新文件名（含路径），如 "docs/note.md"。
	Filename string
	// Content 文件内容（表单 value）。
	Content string
	// Message 提交信息（标题行）。
	Message string
	// Description 扩展提交说明；可空。
	Description string
	// BaseCommit 父提交 OID（表单 commit），来自 GetImplicitContext 的 CommitOID，
	// 作乐观并发校验。空则由服务端按分支当前 HEAD 处理。
	BaseCommit string
}

// CreateFileResponse 是创建文件接口的响应。
// 对着 docs/flows_1_requests/16_POST_VKMW57by18-abyuds-create-main.txt。
type CreateFileResponse struct {
	Data struct {
		// Redirect 提交成功后跳转的分支树页 URL。
		Redirect string `json:"redirect"`
		// Message 服务端消息；成功时为 null。
		Message *string `json:"message"`
		// CommitQuorumPollPath 提交仲裁轮询路径（含新 commit sha）。
		CommitQuorumPollPath string `json:"commitQuorumPollPath"`
	} `json:"data"`
}
