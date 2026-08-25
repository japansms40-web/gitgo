// Package runlog 把账号发布事件格式化为运行日志文案，供前端渲染着色。
// 通用的标签/文案格式委托给 dfclientkit/runlog，这里固定业务动词为"发布"。
package runlog

import (
	"dfclientkit/runlog"

	"githubbaidu/internal/accountpublish"
)

type Kind = runlog.Kind

const (
	KindStart   = runlog.KindStart
	KindSuccess = runlog.KindSuccess
	KindFailure = runlog.KindFailure
	KindSwitch  = runlog.KindSwitch
	KindInfo    = runlog.KindInfo
)

// TagFor 返回某类发布事件的日志标签、着色分类，以及正文是否也要跟着上色。
func TagFor(k accountpublish.EventKind) (tag string, kind Kind, highlightMessage bool) {
	return runlog.TagFor(k)
}

// LineFor 把一条发布事件格式化为日志正文（不含时间戳/标签）。
func LineFor(e accountpublish.Event) string {
	return runlog.LineFor(e, "发布")
}
