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

	got, err := ScanPaths([]string{dir})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	var titles []string
	for _, a := range got {
		titles = append(titles, a.Title)
	}
	sort.Strings(titles)
	want := []string{"deep", "hello", "note"} // .png 被忽略，递归子目录
	if len(titles) != len(want) {
		t.Fatalf("扫描结果 = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Errorf("titles[%d] = %q, want %q", i, titles[i], want[i])
		}
	}
}

func TestScanPaths_TitleAndContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "my-post.md"), "body text")
	got, err := ScanPaths([]string{dir})
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%v err=%v", got, err)
	}
	if got[0].Title != "my-post" {
		t.Errorf("Title = %q, want my-post", got[0].Title)
	}
	if string(got[0].Content) != "body text" {
		t.Errorf("Content = %q", got[0].Content)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
