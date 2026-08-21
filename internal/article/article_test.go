package article

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestScanPaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hello.md"), "# hi")
	writeFile(t, filepath.Join(dir, "note.txt"), "plain")
	writeFile(t, filepath.Join(dir, "ignore.png"), "binary")
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	writeFile(t, filepath.Join(sub, "deep.md"), "deep")

	got, err := ScanPaths([]string{dir}, "posts")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var repoPaths []string
	for _, a := range got {
		repoPaths = append(repoPaths, a.RepoPath)
	}
	sort.Strings(repoPaths)
	want := []string{"posts/deep.md", "posts/hello.md", "posts/note.md"} // .txt→.md，.png 被忽略，递归
	if len(repoPaths) != len(want) {
		t.Fatalf("扫描结果 = %v, want %v", repoPaths, want)
	}
	for i := range want {
		if repoPaths[i] != want[i] {
			t.Errorf("repoPaths[%d] = %q, want %q", i, repoPaths[i], want[i])
		}
	}
}

func TestScanPaths_TitleAndContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "my-post.md"), "body text")
	got, err := ScanPaths([]string{dir}, "")
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%v err=%v", got, err)
	}
	if got[0].Title != "my-post" {
		t.Errorf("Title = %q, want my-post", got[0].Title)
	}
	if string(got[0].Content) != "body text" {
		t.Errorf("Content = %q", got[0].Content)
	}
	if got[0].RepoPath != "my-post.md" { // 空目录=仓库根
		t.Errorf("RepoPath = %q, want my-post.md", got[0].RepoPath)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
