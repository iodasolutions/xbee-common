package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// CopyDir copies src directory into dst, recursively.
// It preserves file modes, and attempts to preserve uid/gid + timestamps.
// On systems where uid/gid is not available (e.g., Windows), it only preserves modes.
// If strictOwner is true, a chown failure aborts the copy.
func CopyDir(src, dst string, strictOwner bool) *cmd.XbeeError {
	if runtime.GOOS == "windows" {
		return cmd.Error("Copy is not supported on Windows")
	}
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Lstat(src)
	if err != nil {
		return cmd.Error("lstat src: %w", err)
	}
	if !srcInfo.IsDir() {
		return cmd.Error("src is not a directory: %s", src)
	}

	// Create root dst dir
	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return cmd.Error("mkdir dst: %w", err)
	}

	// Preserve owner + times on root directory
	if err := preserveMeta(dst, srcInfo, strictOwner); err != nil {
		return cmd.Error("%v", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return cmd.Error("readdir src: %w", err)
	}

	for _, entry := range entries {
		sPath := filepath.Join(src, entry.Name())
		dPath := filepath.Join(dst, entry.Name())

		info, err := os.Lstat(sPath)
		if err != nil {
			return cmd.Error("lstat %s: %w", sPath, err)
		}

		switch {
		case info.IsDir():
			if err := CopyDir(sPath, dPath, strictOwner); err != nil {
				return cmd.Error("%v", err)
			}

		case (info.Mode() & os.ModeSymlink) != 0:
			// Copy symlink as symlink (do not dereference)
			linkTarget, err := os.Readlink(sPath)
			if err != nil {
				return cmd.Error("readlink %s: %w", sPath, err)
			}
			_ = os.RemoveAll(dPath) // in case exists
			if err := os.Symlink(linkTarget, dPath); err != nil {
				return cmd.Error("symlink %s -> %s: %w", dPath, linkTarget, err)
			}
			// For symlink: we can only preserve owner on some Unix via Lchown; mode/time mostly not portable.
			if err := preserveSymlinkOwner(dPath, info, strictOwner); err != nil {
				return cmd.Error("%v", err)
			}

		default:
			if err := copyFile(sPath, dPath, info); err != nil {
				return err
			}
			if err := preserveMeta(dPath, info, strictOwner); err != nil {
				return cmd.Error("%v", err)
			}
		}
	}

	// Finally, re-apply directory mode/meta after children (some ops might affect it)
	if err := os.Chmod(dst, srcInfo.Mode().Perm()); err != nil {
		return cmd.Error("chmod dir %s: %w", dst, err)
	}
	if err := preserveMeta(dst, srcInfo, strictOwner); err != nil {
		return cmd.Error("%v", err)
	}

	return nil
}

func copyFile(src, dst string, info fs.FileInfo) *cmd.XbeeError {
	// Ensure parent exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return cmd.Error("mkdir parent %s: %w", dst, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return cmd.Error("open src %s: %w", src, err)
	}
	defer in.Close()

	// Create with same perms (we’ll chmod again after write to be safe)
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return cmd.Error("create dst %s: %w", dst, err)
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return cmd.Error("copy %s -> %s: %w", src, dst, copyErr)
	}
	if closeErr != nil {
		return cmd.Error("close dst %s: %w", dst, closeErr)
	}

	// Preserve mode including special bits (setuid/setgid/sticky) when possible
	if err := os.Chmod(dst, info.Mode()); err != nil {
		return cmd.Error("chmod %s: %w", dst, err)
	}
	return nil
}

func preserveMeta(path string, srcInfo fs.FileInfo, strictOwner bool) error {
	// 1) Ownership (uid/gid) best effort on Unix
	if err := preserveOwner(path, srcInfo, strictOwner); err != nil {
		return err
	}

	// 2) Timestamps best effort (atime/mtime)
	atime, mtime := fileTimes(srcInfo)
	// On Windows, Chtimes works; on symlinks, it follows link (so we avoided calling it for symlinks)
	if err := os.Chtimes(path, atime, mtime); err != nil {
		// Not fatal in many environments; choose your policy.
		// Here: return error because user asked to preserve; adjust if you want best effort.
		return fmt.Errorf("chtimes %s: %w", path, err)
	}

	// 3) Permissions
	if err := os.Chmod(path, srcInfo.Mode()); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}

	return nil
}
