package ui

import (
	"errors"
	"testing"

	"githubbaidu/internal/publisher"
)

func TestApplyEvent(t *testing.T) {
	st := NewRowState(3)

	st.Apply(publisher.Event{Kind: publisher.EventStart, Index: 0, Title: "a"})
	if st.Status(0) != "发布中" {
		t.Errorf("Start 后状态 = %q", st.Status(0))
	}

	st.Apply(publisher.Event{Kind: publisher.EventSuccess, Index: 0, Title: "a", URL: "http://x"})
	if st.Status(0) != "成功" || st.URL(0) != "http://x" {
		t.Errorf("Success 后 status=%q url=%q", st.Status(0), st.URL(0))
	}
	if st.Done() != 1 {
		t.Errorf("完成计数 = %d, want 1", st.Done())
	}

	st.Apply(publisher.Event{Kind: publisher.EventFailure, Index: 1, Title: "b", Err: errors.New("boom")})
	if st.Status(1) != "失败: boom" {
		t.Errorf("Failure 后状态 = %q", st.Status(1))
	}
	if st.Done() != 2 {
		t.Errorf("失败也计入完成, done=%d", st.Done())
	}

	st.Apply(publisher.Event{Kind: publisher.EventRetry, Index: 2, Title: "c", Err: errors.New("net")})
	if st.Status(2) != "重试中: net" {
		t.Errorf("Retry 后状态 = %q", st.Status(2))
	}
	if st.Done() != 2 {
		t.Errorf("重试不计完成, done=%d", st.Done())
	}
}
