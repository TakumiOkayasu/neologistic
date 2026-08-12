package freeze

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/pathguard"
)

type FileDigest struct {
	SHA256 string `json:"sha256"`
	Mode   string `json:"mode"`
}

type Manifest struct {
	Version string                `json:"version"`
	Files   map[string]FileDigest `json:"files"`
}

// These inputs collectively define what the paid run sees and how it is scored.
// Changing any of them requires a new freeze before another matched run.
var roots = []string{
	"README.md",
	"SPEC.md",
	"INTENT_GATE.md",
	"Makefile",
	"go.mod",
	"cmd",
	"docs",
	"fixtures",
	"internal",
	"prompts",
	"schemas",
	"scripts",
}

func Build(root string) (Manifest, error) {
	canonicalRoot, err := pathguard.CanonicalRoot(root)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{Version: model.FreezeVersion, Files: map[string]FileDigest{}}
	for _, rel := range roots {
		regular, regularErr := pathguard.ResolveRegular(canonicalRoot, rel)
		if regularErr == nil {
			if err := add(canonicalRoot, regular, manifest.Files); err != nil {
				return Manifest{}, err
			}
			continue
		}
		directory, dirErr := pathguard.ResolveDirectory(canonicalRoot, rel)
		if dirErr != nil {
			return Manifest{}, fmt.Errorf("freeze input %s: file error: %v; directory error: %w", rel, regularErr, dirErr)
		}
		err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
				return fmt.Errorf("freeze input is not a regular non-symlink file: %s", path)
			}
			return add(canonicalRoot, path, manifest.Files)
		})
		if err != nil {
			return Manifest{}, err
		}
	}
	return manifest, nil
}

func add(root, path string, files map[string]FileDigest) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "../") || rel == ".." {
		return fmt.Errorf("freeze path escapes root: %s", rel)
	}
	if _, duplicate := files[rel]; duplicate {
		return fmt.Errorf("duplicate freeze path: %s", rel)
	}
	sum := sha256.Sum256(data)
	files[rel] = FileDigest{SHA256: hex.EncodeToString(sum[:]), Mode: fmt.Sprintf("%04o", info.Mode().Perm())}
	return nil
}

func Encode(manifest Manifest) ([]byte, error) {
	if manifest.Version != model.FreezeVersion {
		return nil, fmt.Errorf("unsupported freeze version %q", manifest.Version)
	}
	keys := make([]string, 0, len(manifest.Files))
	for key := range manifest.Files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("{\n  \"version\": ")
	version, _ := json.Marshal(manifest.Version)
	b.Write(version)
	b.WriteString(",\n  \"files\": {\n")
	for i, key := range keys {
		keyJSON, _ := json.Marshal(key)
		valueJSON, err := json.Marshal(manifest.Files[key])
		if err != nil {
			return nil, err
		}
		b.WriteString("    ")
		b.Write(keyJSON)
		b.WriteString(": ")
		b.Write(valueJSON)
		if i+1 < len(keys) {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("  }\n}\n")
	return []byte(b.String()), nil
}

func Load(path string) (Manifest, error) {
	return codec.LoadJSON[Manifest](path)
}

func Verify(root string, expected Manifest) error {
	actual, err := Build(root)
	if err != nil {
		return err
	}
	if expected.Version != actual.Version {
		return fmt.Errorf("freeze version mismatch: expected %s, actual %s", expected.Version, actual.Version)
	}
	var errors []string
	for path, expectedDigest := range expected.Files {
		actualDigest, ok := actual.Files[path]
		if !ok {
			errors = append(errors, "missing: "+path)
			continue
		}
		if actualDigest.SHA256 != expectedDigest.SHA256 {
			errors = append(errors, "changed content: "+path)
		}
		if actualDigest.Mode != expectedDigest.Mode {
			errors = append(errors, "changed mode: "+path)
		}
	}
	for path := range actual.Files {
		if _, ok := expected.Files[path]; !ok {
			errors = append(errors, "unfrozen: "+path)
		}
	}
	sort.Strings(errors)
	if len(errors) != 0 {
		return fmt.Errorf("freeze verification failed: %s", strings.Join(errors, "; "))
	}
	return nil
}
