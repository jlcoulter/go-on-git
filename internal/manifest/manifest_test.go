package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func writeManifest(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadSaveRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, `version = 1
base_url = "https://github.com/jlcoulter"

[[repo]]
path = "momus"

[[repo]]
path = "vendored/other"
url = "https://gitlab.com/foo/other"
`)
	m, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.BaseURL != "https://github.com/jlcoulter" {
		t.Errorf("BaseURL = %q", m.BaseURL)
	}
	if len(m.Repos) != 2 {
		t.Fatalf("len(Repos) = %d", len(m.Repos))
	}
	if m.Repos[0].Path != "momus" || m.Repos[0].URL != "" {
		t.Errorf("repo0 = %+v", m.Repos[0])
	}
	if m.Repos[1].Path != "vendored/other" || m.Repos[1].URL != "https://gitlab.com/foo/other" {
		t.Errorf("repo1 = %+v", m.Repos[1])
	}

	if err := m.Save(path); err != nil {
		t.Fatal(err)
	}
	m2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m2.Repos) != 2 || m2.Repos[1].URL != "https://gitlab.com/foo/other" {
		t.Errorf("roundtrip mismatch: %+v", m2.Repos)
	}
}

func TestFindWalkUp(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, root, "version = 1\n")

	mp, rd, err := Find(sub, "")
	if err != nil {
		t.Fatal(err)
	}
	if rd != root {
		t.Errorf("rootDir = %q, want %q", rd, root)
	}
	if filepath.Dir(mp) != root {
		t.Errorf("manifest dir = %q", filepath.Dir(mp))
	}
}

func TestFindExplicit(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, "version = 1\n")
	mp, rd, err := Find("", path)
	if err != nil {
		t.Fatal(err)
	}
	if mp != path || rd != dir {
		t.Errorf("mp=%q rd=%q", mp, rd)
	}
}

func TestFindMissing(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := Find(dir, ""); err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestResolve(t *testing.T) {
	m := &Manifest{BaseURL: "https://github.com/jlcoulter"}
	u, err := m.Resolve(Repo{Path: "momus"})
	if err != nil || u != "https://github.com/jlcoulter/momus" {
		t.Errorf("derived = %q, %v", u, err)
	}
	u, err = m.Resolve(Repo{Path: "a/b", URL: "https://x/y"})
	if err != nil || u != "https://x/y" {
		t.Errorf("explicit = %q, %v", u, err)
	}
	if _, err := (&Manifest{}).Resolve(Repo{Path: "momus"}); err == nil {
		t.Error("expected error when base_url empty and no url")
	}
}

func TestUpsert(t *testing.T) {
	m := &Manifest{Repos: []Repo{{Path: "b"}}}
	added := m.Upsert([]Repo{{Path: "a"}, {Path: "b"}, {Path: "c"}})
	if len(added) != 2 {
		t.Fatalf("added = %d", len(added))
	}
	if len(m.Repos) != 3 || m.Repos[0].Path != "a" || m.Repos[1].Path != "b" || m.Repos[2].Path != "c" {
		t.Errorf("repos = %+v", m.Repos)
	}
}

func TestRemove(t *testing.T) {
	m := &Manifest{Repos: []Repo{{Path: "a"}, {Path: "b"}}}
	removed, unknown := m.Remove([]string{"a", "zzz"})
	if len(removed) != 1 || removed[0].Path != "a" {
		t.Errorf("removed = %+v", removed)
	}
	if len(unknown) != 1 || unknown[0] != "zzz" {
		t.Errorf("unknown = %+v", unknown)
	}
	if len(m.Repos) != 1 || m.Repos[0].Path != "b" {
		t.Errorf("repos = %+v", m.Repos)
	}
}

func TestValidateRejectsAbsolute(t *testing.T) {
	dir := t.TempDir()
	path := writeManifest(t, dir, "version = 1\n\n[[repo]]\npath = \"/abs\"\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for absolute path")
	}
}

func TestScan(t *testing.T) {
	root := t.TempDir()
	mk := func(p string) {
		if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mkGit := func(p string) {
		mk(p)
		if err := os.MkdirAll(filepath.Join(root, p, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mkGit("repoA")
	mkGit("repoB")
	mkGit("repoB/nested") // nested under a repo: pruned, not discovered
	mk("plain")
	mk("plain/.git") // a dir containing .git is itself a repo root
	// root itself is a repo: skipped
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// worktree: .git is a file
	mk("worktree")
	if err := os.WriteFile(filepath.Join(root, "worktree", ".git"), []byte("gitdir: x"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"plain", "repoA", "repoB", "worktree"}
	if len(paths) != len(want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}
