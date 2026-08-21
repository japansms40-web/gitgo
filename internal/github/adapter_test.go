package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"githubbaidu/internal/publisher"
)

func TestAdapterImplementsInterface(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"content":{"html_url":"u"}}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)

	var pc publisher.GitHubClient = NewAdapter(c) // 编译期确认满足接口
	url, err := pc.PutFile(context.Background(), publisher.PutParams{
		Owner: "a", Repo: "b", Path: "p.md", Branch: "main", Message: "m", Content: []byte("x"),
	})
	if err != nil || url != "u" {
		t.Fatalf("PutFile via adapter: url=%q err=%v", url, err)
	}
}
