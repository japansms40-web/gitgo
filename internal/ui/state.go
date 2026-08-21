package ui

import (
	"fmt"

	"githubbaidu/internal/publisher"
)

// RowState 跟踪每篇文章的显示状态与整体进度。非并发安全，
// 必须只在 Fyne 主线程（fyne.Do 内）调用。
type RowState struct {
	status []string
	urls   []string
	done   int
}

// NewRowState 初始化 n 行，全部为"待发布"。
func NewRowState(n int) *RowState {
	s := &RowState{status: make([]string, n), urls: make([]string, n)}
	for i := range s.status {
		s.status[i] = "待发布"
	}
	return s
}

// Apply 根据事件更新状态。
func (s *RowState) Apply(e publisher.Event) {
	if e.Index < 0 || e.Index >= len(s.status) {
		return
	}
	switch e.Kind {
	case publisher.EventStart:
		s.status[e.Index] = "发布中"
	case publisher.EventSuccess:
		s.status[e.Index] = "成功"
		s.urls[e.Index] = e.URL
		s.done++
	case publisher.EventFailure:
		s.status[e.Index] = "失败: " + errText(e.Err)
		s.done++
	case publisher.EventRetry:
		s.status[e.Index] = "重试中: " + errText(e.Err)
	}
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Status 返回第 i 行状态文本。
func (s *RowState) Status(i int) string { return s.status[i] }

// URL 返回第 i 行成功链接（未成功为空）。
func (s *RowState) URL(i int) string { return s.urls[i] }

// Done 返回已完成（成功+失败）篇数。
func (s *RowState) Done() int { return s.done }

// SuccessURLs 返回所有成功链接，供导出。
func (s *RowState) SuccessURLs() []string {
	var out []string
	for _, u := range s.urls {
		if u != "" {
			out = append(out, u)
		}
	}
	return out
}

// Progress 返回 "done/total" 文本。
func (s *RowState) Progress() string {
	return fmt.Sprintf("%d/%d", s.done, len(s.status))
}
