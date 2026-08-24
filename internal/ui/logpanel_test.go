package ui

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2/theme"

	"githubbaidu/internal/publisher"
)

func TestLogTagColors(t *testing.T) {
	th := NewTheme(theme.VariantLight)

	tag, tagColor, _ := logTag(th, publisher.EventSuccess)
	if tag != "[成功]" || tagColor != th.successColor() {
		t.Errorf("success tag = %q/%#v", tag, tagColor)
	}

	tag, tagColor, msgColor := logTag(th, publisher.EventFailure)
	if tag != "[失败]" || tagColor != th.errorColor() || msgColor != th.errorColor() {
		t.Errorf("failure tag = %q/%#v/%#v", tag, tagColor, msgColor)
	}

	tag, tagColor, _ = logTag(th, publisher.EventRetry)
	if tag != "[重试]" || tagColor != th.warningColor() {
		t.Errorf("retry tag = %q/%#v", tag, tagColor)
	}

	tag, tagColor, _ = logTag(th, publisher.EventStart)
	if tag != "[开始]" || tagColor != th.accentColor() {
		t.Errorf("start tag = %q/%#v", tag, tagColor)
	}
}

func TestLogLineFormatting(t *testing.T) {
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
		if got := logLine(c.event); got != c.want {
			t.Errorf("logLine(%+v) = %q, want %q", c.event, got, c.want)
		}
	}
}
