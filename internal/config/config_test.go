package config

import (
	"testing"

	"gitmd/internal/contentgen"
)

func TestDefaultsAreValid(t *testing.T) {
	got := defaults()
	want := got
	want.Normalize()
	if got != want {
		t.Errorf("默认值本身应当已经是合法的：defaults() = %+v，Normalize 后 = %+v", got, want)
	}
	if got.Count < 1 {
		t.Errorf("默认生成篇数应至少为 1，得到 %d", got.Count)
	}
	if got.KeywordOrder != contentgen.OrderSequential {
		t.Errorf("默认关键词调用方式 = %q，期望 %q", got.KeywordOrder, contentgen.OrderSequential)
	}
}
