package net2

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/util"
	"github.com/jlaffaye/ftp"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const bytesToMegaBytes = 1048576.0

func DownloadIfNotCached(ctx context.Context, rawUrl string) (newfs.File, *util.XbeeError) {
	f := newfs.XbeeIntern().CachedFileForUrl(rawUrl)
	if !f.Exists() {
		err := DoDownload(ctx, rawUrl)
		return f, err
	} else {
		log2.Debugf("Found %s in xbee cache\n", f.Base())
	}
	return f, nil
}
func DownloadIfNotCachedWithWget(ctx context.Context, rawUrl string) (newfs.File, *util.XbeeError) {
	f := newfs.XbeeIntern().CachedFileForUrl(rawUrl)
	if !f.Exists() {
		err := DoDownloadWget(ctx, rawUrl)
		return f, err
	} else {
		log2.Debugf("Found %s in xbee cache\n", f.Base())
	}
	return f, nil
}

func DoDownload(ctx context.Context, rawUrl string) *util.XbeeError {

	anUrl, err := url.Parse(rawUrl)
	if err != nil {
		return util.Error("cannot parse %s : %v", rawUrl, err)
	}
	var pt *PassThru
	if anUrl.Scheme == "ftp" {
		portS := anUrl.Port()
		if portS == "" {
			portS = "21"
		}
		c, err := ftp.Dial(fmt.Sprintf("%s:%s", anUrl.Host, portS), ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			return util.Error("cannot connect to %s : %v", anUrl.Host, err)
		}
		user := anUrl.User.Username()
		pass, _ := anUrl.User.Password()
		if user == "" {
			user = "anonymous"
			pass = "anonymous"
		}
		err = c.Login(user, pass)
		if err != nil {
			return util.Error("login to %s failed : %v", anUrl.Host, err)
		}
		aPath := newfs.Path(anUrl.Path)
		entries, err := c.List(aPath.Dir().String())
		if err != nil {
			return util.Error("cannot list content of host %s, path %s : %v", anUrl.Host, aPath.Dir(), err)
		}
		var size int64
		for _, anEntry := range entries {
			if aPath.Base() == anEntry.Name {
				size = int64(anEntry.Size)
				break
			}
		}
		r, err := c.Retr(anUrl.Path)

		if err != nil {
			return util.Error("%v", err)
		}
		defer r.Close()
		pt = NewPathThru(r, size)
	} else { // default is http or https
		resp, err := http.Get(rawUrl)
		if err != nil {
			return util.Error("Failed invoke http GET on url %s : %v\n", rawUrl, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return util.Error("Server responded with status code = %d for url %s", resp.StatusCode, rawUrl)
		}
		pt = NewPathThru(resp.Body, resp.ContentLength)
	}
	f := newfs.XbeeIntern().CachedFileForUrl(rawUrl)
	if hUser := HostUser(); hUser != nil {
		if err := hUser.ChangeOwnerForAnscestorsOf(f, newfs.XbeeIntern().CacheArtefacts()); err != nil {
			return err
		}
	}
	return pt.DownloadTo(f)
}
func DoDownloadWget(ctx context.Context, rawUrl string) *util.XbeeError {
	f := newfs.XbeeIntern().CachedFileForUrl(rawUrl)
	fd := f.Dir()
	fd.EnsureExists()
	execC := exec.CommandContext(ctx, "wget", rawUrl)
	log2.Debugf("DoDownloadWget : %s ", execC.String())
	execC.Stderr = os.Stderr
	execC.Stdout = os.Stdout
	execC.Stdin = os.Stdin
	execC.Dir = fd.String()
	if err := execC.Run(); err != nil {
		return util.Error("command %v failed : %v", execC.String(), err)
	}
	if hUser := HostUser(); hUser != nil {
		uid, _ := strconv.Atoi(hUser.Uid())
		gid, _ := strconv.Atoi(hUser.Gid())
		if err := os.Chown(f.String(), uid, gid); err != nil {
			return util.Error("Cannot set ownership to %s:%s for file %s", hUser.Uid(), hUser.Gid(), f)
		}
	}
	return nil
}

type PassThru struct {
	io.Reader
	curr  int64
	total int64
	start time.Time
}

func NewPathThru(r io.Reader, length int64) *PassThru {
	return &PassThru{
		Reader: r,
		total:  length,
		start:  time.Now(),
	}
}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.curr += int64(n)

	// last read will have EOF err
	if err == nil || (err == io.EOF && n > 0) {
		pt.printProgress(pt.curr, pt.total)
	}
	return n, err
}

func (pt *PassThru) printProgress(curr, total int64) {
	width := 40.0
	output := ""
	threshold := (float64(curr) / float64(total)) * width
	for i := 0.0; i < width; i++ {
		if i < threshold {
			if output == "" {
				output = ">"
			} else {
				output = "=" + output
			}
		} else {
			output += " "
		}
	}
	perc := (float64(curr) / float64(total)) * 100
	duree := time.Now().Sub(pt.start)
	message := fmt.Sprintf("\r%3.0f%%[%s] %.1fMB %.2fMB/s eta %v", perc, output, float64(curr)/bytesToMegaBytes, float64(curr)/bytesToMegaBytes/duree.Seconds(), duree.Round(time.Second))
	fmt.Print(message)
}

func (pt *PassThru) DownloadTo(f newfs.File) *util.XbeeError {
	tmpFile := newfs.File(f.String() + ".tmp")
	size := tmpFile.FillFromReader(pt)
	fmt.Print("\n")
	if err := os.Rename(tmpFile.String(), f.String()); err != nil {
		return util.Error("cannot rename %s to %s", tmpFile, f)
	}
	if hUser := HostUser(); hUser != nil {
		uid, _ := strconv.Atoi(hUser.Uid())
		gid, _ := strconv.Atoi(hUser.Gid())
		if err := os.Chown(f.String(), uid, gid); err != nil {
			return util.Error("Cannot set ownership to %s:%s for file %s", hUser.Uid(), hUser.Gid(), f)
		}
	}
	log2.Infof("Resource %s Transferred. (%.1f MB)\n", f.Base(), float64(size)/bytesToMegaBytes)
	return nil
}
