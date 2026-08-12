package artifact

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
)

const Version = "phase3-artifact-manifest-v1"

type FileDigest struct {
	SHA256 string `json:"sha256"`
	Mode   string `json:"mode"`
	Size   int64  `json:"size"`
}

type Manifest struct {
	Version string                `json:"version"`
	Files   map[string]FileDigest `json:"files"`
}

func Build(directory, excludedRelativePath string) (Manifest, error) {
	root, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return Manifest{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return Manifest{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Manifest{}, err
	}
	if !info.IsDir() {
		return Manifest{}, fmt.Errorf("artifact root is not a directory: %s", directory)
	}
	excluded := filepath.ToSlash(filepath.Clean(filepath.FromSlash(excludedRelativePath)))
	if excluded == "." {
		excluded = ""
	}
	manifest := Manifest{Version: Version, Files: map[string]FileDigest{}}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fmt.Errorf("artifact input is not a regular non-symlink file: %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == excluded {
			return nil
		}
		if strings.HasPrefix(rel, "../") || rel == ".." {
			return fmt.Errorf("artifact path escapes root: %s", rel)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		manifest.Files[rel] = FileDigest{
			SHA256: hex.EncodeToString(sum[:]),
			Mode:   fmt.Sprintf("%04o", info.Mode().Perm()),
			Size:   info.Size(),
		}
		return nil
	})
	if err != nil {
		return Manifest{}, err
	}
	if len(manifest.Files) == 0 {
		return Manifest{}, fmt.Errorf("artifact manifest would be empty")
	}
	return manifest, nil
}

func Encode(manifest Manifest) ([]byte, error) {
	if manifest.Version != Version {
		return nil, fmt.Errorf("unsupported artifact manifest version %q", manifest.Version)
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

func Decode(data []byte) (Manifest, error) {
	manifest, err := codec.DecodeStrict[Manifest](data)
	if err != nil {
		return Manifest{}, err
	}
	if manifest.Version != Version {
		return Manifest{}, fmt.Errorf("unsupported artifact manifest version %q", manifest.Version)
	}
	if len(manifest.Files) == 0 {
		return Manifest{}, fmt.Errorf("artifact manifest is empty")
	}
	return manifest, nil
}

func Verify(directory, excludedRelativePath string, expected Manifest) error {
	actual, err := Build(directory, excludedRelativePath)
	if err != nil {
		return err
	}
	var errors []string
	for path, want := range expected.Files {
		got, ok := actual.Files[path]
		if !ok {
			errors = append(errors, "missing: "+path)
			continue
		}
		if got.SHA256 != want.SHA256 {
			errors = append(errors, "changed content: "+path)
		}
		if got.Mode != want.Mode {
			errors = append(errors, "changed mode: "+path)
		}
		if got.Size != want.Size {
			errors = append(errors, "changed size: "+path)
		}
	}
	for path := range actual.Files {
		if _, ok := expected.Files[path]; !ok {
			errors = append(errors, "unmanifested: "+path)
		}
	}
	sort.Strings(errors)
	if len(errors) != 0 {
		return fmt.Errorf("artifact verification failed: %s", strings.Join(errors, "; "))
	}
	return nil
}
