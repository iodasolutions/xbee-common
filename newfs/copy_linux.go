package newfs

import (
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"syscall"
	"time"
)

func preserveOwner(path string, srcInfo fs.FileInfo, strictOwner bool) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	stat, ok := srcInfo.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return nil
	}
	if err := os.Chown(path, int(stat.Uid), int(stat.Gid)); err != nil {
		if strictOwner {
			return fmt.Errorf("chown %s: %w", path, err)
		}
		// best-effort: ignore permission errors etc.
	}
	return nil
}

func fileTimes(info fs.FileInfo) (atime time.Time, mtime time.Time) {
	mtime = info.ModTime()

	// Try to get atime from syscall.Stat_t on Unix.
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat != nil {
		// Linux: Atim; Darwin/FreeBSD: Atimespec
		// We handle common cases with build-agnostic reflection via fields we know.
		// If not available, fall back to mtime.
		// Linux:
		//   stat.Atim.Sec / stat.Atim.Nsec
		// macOS:
		//   stat.Atimespec.Sec / stat.Atimespec.Nsec
		type atimLike struct{ Sec, Nsec int64 }

		// Try linux-style
		if v, ok := any(stat).(interface{ GetAtim() atimLike }); ok {
			a := v.GetAtim()
			return time.Unix(a.Sec, a.Nsec), mtime
		}

		// Direct field access for common platforms:
		// We do simple best-effort with what exists in stdlib structs:
		// On Linux, Stat_t has Atim; on Darwin, it has Atimespec.
		// We can’t portably access those without build tags,
		// so we keep atime = mtime if we can't.
	}

	return mtime, mtime
}

func preserveSymlinkOwner(path string, srcInfo fs.FileInfo, strictOwner bool) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	stat, ok := srcInfo.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return nil
	}
	// Lchown changes link ownership instead of target
	if err := os.Lchown(path, int(stat.Uid), int(stat.Gid)); err != nil {
		if strictOwner {
			return fmt.Errorf("lchown symlink %s: %w", path, err)
		}
	}
	return nil
}
