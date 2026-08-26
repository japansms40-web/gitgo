package main

import (
	"errors"
	"testing"
)

// TestAccountVerdict 验证账号验活的判定逻辑（纯函数，不联网）：
// 成功=活号；401/403=坏号；其它错误=失败。
func TestAccountVerdict(t *testing.T) {
	someErr := errors.New("网络错误")

	t.Run("成功即活号", func(t *testing.T) {
		r := accountVerdict(3, 200, nil)
		if !r.Ok || r.Bad {
			t.Fatalf("期望 Ok=true Bad=false, 得到 %+v", r)
		}
		if r.RepoCount != 3 {
			t.Errorf("RepoCount = %d, 期望 3", r.RepoCount)
		}
	})

	t.Run("401 是坏号", func(t *testing.T) {
		r := accountVerdict(0, 401, someErr)
		if r.Ok || !r.Bad {
			t.Fatalf("期望 Bad=true Ok=false, 得到 %+v", r)
		}
	})

	t.Run("403 是坏号", func(t *testing.T) {
		r := accountVerdict(0, 403, someErr)
		if !r.Bad {
			t.Errorf("403 应判坏号, 得到 %+v", r)
		}
	})

	t.Run("网络错误是失败不是坏号", func(t *testing.T) {
		r := accountVerdict(0, 0, someErr)
		if r.Ok {
			t.Errorf("网络错误不应 Ok")
		}
		if r.Bad {
			t.Errorf("网络错误不该判坏号（可能是代理/网络问题）, 得到 %+v", r)
		}
	})
}
