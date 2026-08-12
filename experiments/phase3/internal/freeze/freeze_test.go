package freeze

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func populateFreezeRoot(t *testing.T, root string) {
	t.Helper()
	paths := []string{
		"README.md", "SPEC.md", "INTENT_GATE.md", "Makefile", "go.mod",
		"cmd/main.go", "docs/METHODOLOGY.md", "fixtures/arms.json", "fixtures/cases.jsonl",
		"internal/eval/eval.go", "prompts/a.md", "schemas/a.json", "scripts/check.sh",
	}
	for _, rel := range paths {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if rel == "scripts/check.sh" {
			mode = 0o755
		}
		if err := os.WriteFile(full, []byte(rel+"\n"), mode); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBuildAndVerifyDetectsContentAndModeChange(t *testing.T) {
	root := t.TempDir()
	populateFreezeRoot(t, root)
	manifest, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 13 {
		t.Fatalf("unexpected frozen file count: %d", len(manifest.Files))
	}
	if err := Verify(root, manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "SPEC.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root, manifest); err == nil {
		t.Fatal("expected changed file to invalidate manifest")
	}
	populateFreezeRoot(t, root)
	manifest, err = Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "scripts", "check.sh"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root, manifest); err == nil {
		t.Fatal("expected mode change to invalidate manifest")
	}
}

func TestBuildRejectsSymlinkedInput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	populateFreezeRoot(t, root)
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "SPEC.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "SPEC.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root); err == nil {
		t.Fatal("expected symlinked freeze input rejection")
	}
}
