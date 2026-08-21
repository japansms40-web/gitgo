package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(srv *httptest.Server) *Client {
	c := New("test-token")
	c.baseURL = srv.URL
	c.http = srv.Client()
	return c
}

func TestValidateToken_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("缺少或错误的 Authorization 头: %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/user" {
			t.Errorf("期望 /user, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"login":"alice"}`))
	}))
	defer srv.Close()

	login, err := newTestClient(srv).ValidateToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if login != "alice" {
		t.Errorf("login = %q, want alice", login)
	}
}

func TestValidateToken_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer srv.Close()

	_, err := newTestClient(srv).ValidateToken(context.Background())
	if err == nil {
		t.Fatal("期望返回错误，got nil")
	}
}

func TestRepoExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/alice/exists" {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)

	ok, err := c.RepoExists(context.Background(), "alice", "exists")
	if err != nil || !ok {
		t.Fatalf("exists: ok=%v err=%v", ok, err)
	}
	ok, err = c.RepoExists(context.Background(), "alice", "missing")
	if err != nil || ok {
		t.Fatalf("missing: ok=%v err=%v", ok, err)
	}
}

func TestCreateRepo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/user/repos" {
			t.Errorf("期望 POST /user/repos, got %s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		if body["name"] != "newrepo" || body["auto_init"] != true {
			t.Errorf("body 不正确: %v", body)
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()

	if err := newTestClient(srv).CreateRepo(context.Background(), "newrepo"); err != nil {
		t.Fatalf("CreateRepo err: %v", err)
	}
}

func TestGetFileSHA(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/alice/blog/contents/posts/a.md" {
			if r.URL.Query().Get("ref") != "main" {
				t.Errorf("期望 ref=main, got %q", r.URL.Query().Get("ref"))
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"sha":"abc123"}`))
			return
		}
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)

	sha, err := c.GetFileSHA(context.Background(), "alice", "blog", "posts/a.md", "main")
	if err != nil || sha != "abc123" {
		t.Fatalf("存在文件: sha=%q err=%v", sha, err)
	}
	sha, err = c.GetFileSHA(context.Background(), "alice", "blog", "posts/none.md", "main")
	if err != nil || sha != "" {
		t.Fatalf("不存在文件应返回空 sha 且无错误: sha=%q err=%v", sha, err)
	}
}

func TestPutFile_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("期望 PUT, got %s", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		if _, hasSHA := body["sha"]; hasSHA {
			t.Errorf("新建文件不应带 sha")
		}
		if body["content"] == "" {
			t.Errorf("content 不应为空")
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"content":{"html_url":"https://github.com/alice/blog/blob/main/posts/a.md"}}`))
	}))
	defer srv.Close()

	url, err := newTestClient(srv).PutFile(context.Background(), PutParams{
		Owner: "alice", Repo: "blog", Path: "posts/a.md", Branch: "main",
		Message: "add a.md", Content: []byte("hello"), SHA: "",
	})
	if err != nil {
		t.Fatalf("PutFile err: %v", err)
	}
	if url != "https://github.com/alice/blog/blob/main/posts/a.md" {
		t.Errorf("html_url = %q", url)
	}
}

func TestPutFile_RateLimitRetry(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"content":{"html_url":"u"}}`))
	}))
	defer srv.Close()

	_, err := newTestClient(srv).PutFile(context.Background(), PutParams{
		Owner: "a", Repo: "b", Path: "p.md", Branch: "main", Message: "m", Content: []byte("x"),
	})
	if err != nil {
		t.Fatalf("限流后应重试成功, err=%v", err)
	}
	if calls != 2 {
		t.Errorf("期望重试后共 2 次调用, got %d", calls)
	}
}
