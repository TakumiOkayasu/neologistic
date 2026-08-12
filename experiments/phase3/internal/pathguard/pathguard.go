package pathguard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CanonicalRoot resolves root and all filesystem aliases before later
// containment checks. Both operands are canonicalized, avoiding macOS
// /var versus /private/var mismatches.
func CanonicalRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("root must not be empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("root is not a directory: %s", resolved)
	}
	return filepath.Clean(resolved), nil
}

func ResolveRegular(root, rel string) (string, error) {
	return resolveExisting(root, rel, false)
}

func ResolveDirectory(root, rel string) (string, error) {
	return resolveExisting(root, rel, true)
}

func resolveExisting(root, rel string, directory bool) (string, error) {
	if strings.TrimSpace(rel) == "" || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path must be non-empty and relative: %q", rel)
	}
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %q", rel)
	}
	canonicalRoot, err := CanonicalRoot(root)
	if err != nil {
		return "", err
	}
	joined := filepath.Join(canonicalRoot, clean)
	if info, err := os.Lstat(joined); err != nil {
		return "", err
	} else if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("final path component must not be a symlink: %q", rel)
	}
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", err
	}
	if err := ensureWithin(canonicalRoot, resolved); err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if directory && !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", clean)
	}
	if !directory && !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", clean)
	}
	return filepath.Clean(resolved), nil
}

func JoinForCreate(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path must be non-empty and relative: %q", rel)
	}
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %q", rel)
	}
	canonicalRoot, err := CanonicalRoot(root)
	if err != nil {
		return "", err
	}
	parent := filepath.Dir(filepath.Join(canonicalRoot, clean))
	resolvedParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return "", err
	}
	if err := ensureWithin(canonicalRoot, resolvedParent); err != nil {
		return "", err
	}
	target := filepath.Join(resolvedParent, filepath.Base(clean))
	info, err := os.Lstat(target)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("output path must not be a symlink: %q", rel)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("output path must be a regular file or absent: %q", rel)
		}
	case os.IsNotExist(err):
		// Safe to create below the already canonicalized parent.
	case err != nil:
		return "", err
	}
	return target, nil
}

func ensureWithin(root, target string) error {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("path is outside root: %s", target)
	}
	return nil
}
