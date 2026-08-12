package artifact

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildEncodeDecodeAndVerify(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cells", "p3-001", "A"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "run-manifest.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cells", "p3-001", "A", "response.txt"), []byte("response\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, err := Build(root, "artifact-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	data, err := Encode(manifest)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(root, "artifact-manifest.json", decoded); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cells", "p3-001", "A", "response.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root, "artifact-manifest.json", decoded); err == nil {
		t.Fatal("expected artifact mutation rejection")
	}
}

func TestBuildRejectsSymlinkedArtifact(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "response.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root, "artifact-manifest.json"); err == nil {
		t.Fatal("expected symlink rejection")
	}
}
