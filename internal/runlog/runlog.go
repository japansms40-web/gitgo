// Package runlog 把账号发布事件格式化为运行日志文案，供前端渲染着色。
package runlog

import (
	"fmt"

	"githubbaidu/internal/accountpublish"
)

// Kind 标识一条日志在前端应使用的着色分类，前端据此映射到主题色。
type Kind string

const (
	KindStart   Kind = "start"
	KindSuccess Kind = "success"
	KindFailure Kind = "failure"
	KindSwitch  Kind = "switch"
	KindInfo    Kind = "info"
)

// TagFor 返回某类发布事件的日志标签、着色分类，以及正文是否也要跟着上色
// （失败/换号的正文本身也标红/标黄，其余类型正文用默认前景色）。
// EventRoundProgress 不产生日志行，调用方应在此之前过滤掉。
func TagFor(k accountpublish.EventKind) (tag string, kind Kind, highlightMessage bool) {
	switch k {
	case accountpublish.EventAttemptStart:
		return "[开始]", KindStart, false
	case accountpublish.EventAttemptSuccess:
		return "[成功]", KindSuccess, false
	case accountpublish.EventAttemptFailure:
		return "[失败]", KindFailure, true
	case accountpublish.EventAccountSwitch:
		return "[换号]", KindSwitch, true
	case accountpublish.EventRoundStart, accountpublish.EventRoundDone:
		return "[轮次]", KindInfo, false
	default:
		return "[信息]", KindInfo, false
	}
}

// LineFor 把一条发布事件格式化为日志正文（不含时间戳/标签）。
func LineFor(e accountpublish.Event) string {
	switch e.Kind {
	case accountpublish.EventAttemptStart:
		return fmt.Sprintf("账号 %s 开始发布《%s》", e.CK, e.ArticleTitle)
	case accountpublish.EventAttemptSuccess:
		return fmt.Sprintf("账号 %s 发布《%s》成功: %s", e.CK, e.ArticleTitle, e.Result)
	case accountpublish.EventAttemptFailure:
		return fmt.Sprintf("账号 %s 发布《%s》失败: %v", e.CK, e.ArticleTitle, e.Err)
	case accountpublish.EventAccountSwitch:
		return fmt.Sprintf("账号 %s 换号: %v", e.CK, e.Err)
	case accountpublish.EventRoundStart:
		return fmt.Sprintf("第 %d 轮开始，共 %d 个账号", e.Round, e.RoundTotal)
	case accountpublish.EventRoundDone:
		return fmt.Sprintf("第 %d 轮结束", e.Round)
	default:
		return e.CK
	}
}
