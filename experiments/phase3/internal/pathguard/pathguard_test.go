package pathguard

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveRegularWithinRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "prompts", "a.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveRegular(root, "prompts/a.md")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(resolved) != "a.md" {
		t.Fatalf("unexpected resolved path: %s", resolved)
	}
}

func TestResolveRejectsTraversalAndSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "outside.md")
	if err := os.WriteFile(outsideFile, []byte("no\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRegular(root, "../outside.md"); err == nil {
		t.Fatal("expected traversal rejection")
	}
	link := filepath.Join(root, "escape.md")
	if err := os.Symlink(outsideFile, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink unavailable: %v", err)
		}
		t.Fatal(err)
	}
	if _, err := ResolveRegular(root, "escape.md"); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestCanonicalizesBothRootAndTargetAliases(t *testing.T) {
	base := t.TempDir()
	realRoot := filepath.Join(base, "real")
	if err := os.MkdirAll(realRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realRoot, "file.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(realRoot, alias); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink unavailable: %v", err)
		}
		t.Fatal(err)
	}
	if _, err := ResolveRegular(alias, "file.txt"); err != nil {
		t.Fatalf("canonical aliases should resolve consistently: %v", err)
	}
}

func TestJoinForCreateRejectsExistingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "freeze.lock.json")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if _, err := JoinForCreate(root, "freeze.lock.json"); err == nil {
		t.Fatal("expected output symlink rejection")
	}
}
