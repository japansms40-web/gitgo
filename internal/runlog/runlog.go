// Package runlog 把发布事件格式化为运行日志文案，供前端渲染着色。
package runlog

import (
	"fmt"

	"githubbaidu/internal/publisher"
)

// Kind 标识一条日志在前端应使用的着色分类，前端据此映射到主题色。
type Kind string

const (
	KindStart   Kind = "start"
	KindSuccess Kind = "success"
	KindFailure Kind = "failure"
	KindRetry   Kind = "retry"
	KindInfo    Kind = "info"
)

// TagFor 返回某类发布事件的日志标签、着色分类，以及正文是否也要跟着上色
// （失败/重试的正文本身也标红/标黄，其余类型正文用默认前景色）。
func TagFor(k publisher.EventKind) (tag string, kind Kind, highlightMessage bool) {
	switch k {
	case publisher.EventStart:
		return "[开始]", KindStart, false
	case publisher.EventSuccess:
		return "[成功]", KindSuccess, false
	case publisher.EventFailure:
		return "[失败]", KindFailure, true
	case publisher.EventRetry:
		return "[重试]", KindRetry, true
	default:
		return "[信息]", KindInfo, false
	}
}

// LineFor 把一条发布事件格式化为日志正文（不含时间戳/标签）。
func LineFor(e publisher.Event) string {
	switch e.Kind {
	case publisher.EventStart:
		return e.Title
	case publisher.EventSuccess:
		return fmt.Sprintf("%s → %s", e.Title, e.URL)
	case publisher.EventFailure:
		return fmt.Sprintf("%s 失败: %v", e.Title, e.Err)
	case publisher.EventRetry:
		return fmt.Sprintf("%s 重试: %v", e.Title, e.Err)
	default:
		return e.Title
	}
}
