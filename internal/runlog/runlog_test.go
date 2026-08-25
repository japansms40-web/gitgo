package runlog

import (
	"errors"
	"testing"

	"githubbaidu/internal/publisher"
)

func TestTagForKindsAndHighlight(t *testing.T) {
	cases := []struct {
		kind     publisher.EventKind
		wantTag  string
		wantKind Kind
		wantHi   bool
	}{
		{publisher.EventStart, "[开始]", KindStart, false},
		{publisher.EventSuccess, "[成功]", KindSuccess, false},
		{publisher.EventFailure, "[失败]", KindFailure, true},
		{publisher.EventRetry, "[重试]", KindRetry, true},
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
		event publisher.Event
		want  string
	}{
		{publisher.Event{Kind: publisher.EventStart, Title: "hello"}, "hello"},
		{publisher.Event{Kind: publisher.EventSuccess, Title: "hello", URL: "http://x/y"}, "hello → http://x/y"},
		{publisher.Event{Kind: publisher.EventFailure, Title: "hello", Err: errors.New("boom")}, "hello 失败: boom"},
		{publisher.Event{Kind: publisher.EventRetry, Title: "hello", Err: errors.New("net")}, "hello 重试: net"},
	}
	for _, c := range cases {
		if got := LineFor(c.event); got != c.want {
			t.Errorf("LineFor(%+v) = %q, want %q", c.event, got, c.want)
		}
	}
}
