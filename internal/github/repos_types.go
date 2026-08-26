package github

import "encoding/json"

// ReposResponse 是 GitHub 网页版 GET /repos?q=owner:@me&page=N 的响应。
// 字段对着 docs/flows_1_requests/01_GET_repos.txt 抓包的响应体。
type ReposResponse struct {
	Meta    Meta    `json:"meta"`
	Payload Payload `json:"payload"`
}

// Meta 是响应里的页面元信息。
type Meta struct {
	Title string `json:"title"`
}

// Payload 承载业务数据。仓库列表在 reposFinderPageRoute 下。
type Payload struct {
	ReposFinderPageRoute ReposFinderPageRoute `json:"reposFinderPageRoute"`
}

// ReposFinderPageRoute 是仓库列表页的分页数据。
type ReposFinderPageRoute struct {
	PageCount       int          `json:"pageCount"`
	Repositories    []Repository `json:"repositories"`
	RepositoryCount int          `json:"repositoryCount"`
}

// Repository 是一个仓库条目。primaryLanguage / license / participation 抓包为 null 且形状
// 不定，用 json.RawMessage 兜底，避免过度建模。
type Repository struct {
	Type              string          `json:"type"`
	Name              string          `json:"name"`
	Owner             string          `json:"owner"`
	IsFork            bool            `json:"isFork"`
	Description       string          `json:"description"`
	AllTopics         []string        `json:"allTopics"`
	PullRequestCount  int             `json:"pullRequestCount"`
	IssueCount        int             `json:"issueCount"`
	StarsCount        int             `json:"starsCount"`
	ForksCount        int             `json:"forksCount"`
	LastUpdated       LastUpdated     `json:"lastUpdated"`
	IsAdminableByUser bool            `json:"isAdminableByUser"`
	PrimaryLanguage   json.RawMessage `json:"primaryLanguage"`
	License           json.RawMessage `json:"license"`
	Participation     json.RawMessage `json:"participation"`
}

// LastUpdated 是仓库最近一次更新的信息。
type LastUpdated struct {
	HasBeenPushedTo bool   `json:"hasBeenPushedTo"`
	Timestamp       string `json:"timestamp"`
}
