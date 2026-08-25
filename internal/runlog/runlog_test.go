package runlog

import (
	"errors"
	"testing"

	"githubbaidu/internal/accountpublish"
)

func TestTagForKindsAndHighlight(t *testing.T) {
	cases := []struct {
		kind     accountpublish.EventKind
		wantTag  string
		wantKind Kind
		wantHi   bool
	}{
		{accountpublish.EventAttemptStart, "[开始]", KindStart, false},
		{accountpublish.EventAttemptSuccess, "[成功]", KindSuccess, false},
		{accountpublish.EventAttemptFailure, "[失败]", KindFailure, true},
		{accountpublish.EventAccountSwitch, "[换号]", KindSwitch, true},
		{accountpublish.EventRoundStart, "[轮次]", KindInfo, false},
		{accountpublish.EventRoundDone, "[轮次]", KindInfo, false},
	}
	for _, c := range cases {
		tag, kind, hi := TagFor(c.kind)
		if tag != c.wantTag || kind != c.wantKind || hi != c.wantHi {
			t.Errorf("TagFor(%v) = %q/%v/%v, want %q/%v/%v", c.kind, tag, kind, hi, c.wantTag, c.wantKind, c.wantHi)
		}
	}
}

func TestLineForFormatting(t *testing.T) {
	cases := []struct {
		event accountpublish.Event
		want  string
	}{
		{accountpublish.Event{Kind: accountpublish.EventAttemptStart, CK: "ck1", ArticleTitle: "hello"}, "账号 ck1 开始发布《hello》"},
		{accountpublish.Event{Kind: accountpublish.EventAttemptSuccess, CK: "ck1", ArticleTitle: "hello", Result: "ok"}, "账号 ck1 发布《hello》成功: ok"},
		{accountpublish.Event{Kind: accountpublish.EventAttemptFailure, CK: "ck1", ArticleTitle: "hello", Err: errors.New("boom")}, "账号 ck1 发布《hello》失败: boom"},
		{accountpublish.Event{Kind: accountpublish.EventAccountSwitch, CK: "ck1", Err: errors.New("连续失败达到换号阈值")}, "账号 ck1 换号: 连续失败达到换号阈值"},
		{accountpublish.Event{Kind: accountpublish.EventRoundStart, Round: 2, RoundTotal: 5}, "第 2 轮开始，共 5 个账号"},
		{accountpublish.Event{Kind: accountpublish.EventRoundDone, Round: 2}, "第 2 轮结束"},
	}
	for _, c := range cases {
		if got := LineFor(c.event); got != c.want {
			t.Errorf("LineFor(%+v) = %q, want %q", c.event, got, c.want)
		}
	}
}
