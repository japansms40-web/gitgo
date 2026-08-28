package publish

import (
	"math/rand"
	"strings"
	"sync"
	"testing"

	"gitmd/internal/contentgen"
)

// seqLib 是一份只产出 {顺序外链} 的素材，便于直接从标题读出取到了哪条外链。
func seqLib(urls []string) contentgen.Library {
	return contentgen.Library{
		TitleTemplate: "{顺序外链}",
		BodyTemplates: []string{"x"},
		URLs:          urls,
	}
}

// TestRunnerGenBatchSharesCursorAcrossAccounts 一次发布任务里换了账号也要接着往下取外链，
// 而不是每个账号都从外链库第一条重新开始。
func TestRunnerGenBatchSharesCursorAcrossAccounts(t *testing.T) {
	urls := []string{"u0", "u1", "u2", "u3", "u4", "u5"}
	r := New(Config{PerAccount: 2}, nil, seqLib(urls), contentgen.Options{}, nil, 1)

	var got []string
	for account := 0; account < 3; account++ { // 模拟 3 个账号各发 2 篇
		rnd := rand.New(rand.NewSource(int64(account)))
		for _, d := range r.genBatch(rnd, 2) {
			got = append(got, d.Title)
		}
	}
	if want := strings.Join(urls, ","); strings.Join(got, ",") != want {
		t.Errorf("跨账号应顺序消费外链，得到 %v，期望 %v", got, urls)
	}
}

// TestRunnerGenBatchWrapsOnlyAfterExhausted 外链库整份取完一遍才回到第一条。
func TestRunnerGenBatchWrapsOnlyAfterExhausted(t *testing.T) {
	r := New(Config{PerAccount: 1}, nil, seqLib([]string{"u0", "u1", "u2"}), contentgen.Options{}, nil, 1)

	var got []string
	for i := 0; i < 5; i++ {
		for _, d := range r.genBatch(rand.New(rand.NewSource(1)), 1) {
			got = append(got, d.Title)
		}
	}
	if want := "u0,u1,u2,u0,u1"; strings.Join(got, ",") != want {
		t.Errorf("取完一遍才回头，得到 %v，期望 %s", got, want)
	}
}

// TestRunnerGenBatchConcurrentNoDuplicate 多 worker 并发生成时同一条外链不会被两篇用到（配 -race）。
func TestRunnerGenBatchConcurrentNoDuplicate(t *testing.T) {
	const workers, perWorker = 8, 25
	urls := make([]string, workers*perWorker)
	for i := range urls {
		urls[i] = "u" + string(rune('A'+i%26)) + strings.Repeat("x", i/26+1)
	}
	r := New(Config{PerAccount: perWorker}, nil, seqLib(urls), contentgen.Options{}, nil, 1)

	out := make(chan string, workers*perWorker)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for _, d := range r.genBatch(rand.New(rand.NewSource(int64(w))), perWorker) {
				out <- d.Title
			}
		}(w)
	}
	wg.Wait()
	close(out)

	seen := map[string]bool{}
	for u := range out {
		if seen[u] {
			t.Fatalf("外链 %q 被用了两次", u)
		}
		seen[u] = true
	}
	if len(seen) != len(urls) {
		t.Errorf("应把 %d 条外链各用一次，实际用到 %d 条", len(urls), len(seen))
	}
}
