package indus

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/exec2"
	"github.com/iodasolutions/xbee-common/newfs"
	"os"
	"os/exec"
	"strings"
	"time"
)

const publicBucket = "xbee.repository.public"

var PrivateBucket = "xbee.repository"

var osArchs = [][]string{
	{
		"windows", "amd64",
	},
	{
		"linux", "amd64",
	},
	{
		"linux", "arm64",
	},
}

func Build(ctx context.Context, srcMainPath string, execName string) *cmd.XbeeError {
	commit, release, err := CommitAndRelease(ctx)
	if err != nil {
		return err
	}
	for _, osArch := range osArchs {
		if targetBin, err := buildFor(ctx, commit, release, osArch[0], osArch[1], srcMainPath, execName); err != nil {
			return err
		} else {
			copyMaybeToLocalBin(targetBin, osArch[0], osArch[1])
			createArchive(targetBin, osArch[0])
		}
	}
	return nil
}

func createArchive(binFile newfs.File, osName string) (newfs.File, *cmd.XbeeError) {
	if osName == "windows" {
		zFile, err := binFile.Compress("zip")
		if err != nil {
			return "", err
		}
		fmt.Println("zipped OK")
		return zFile, nil
	} else {
		zFile, err := binFile.Compress("gz")
		if err != nil {
			return "", err
		}
		fmt.Println("gz OK")
		return zFile, nil
	}
}

func CommitAndRelease(ctx context.Context) (string, string, *cmd.XbeeError) {
	commit, err := exec2.RunReturnStdOut(ctx, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", "", err
	}
	aCmd := exec2.NewCommand("git", "describe", "--tags", commit).Quiet()
	var release string
	if err := aCmd.Run(ctx); err == nil {
		release = aCmd.Result()
	}
	return strings.TrimSpace(commit), strings.TrimSpace(release), nil
}

func BuildAndDeploy(ctx context.Context, srcMainPath string, execName string) *cmd.XbeeError {
	commit, release, err := CommitAndRelease(ctx)
	if err != nil {
		return err
	}
	svc, err2 := AdminClient(publicBucket)
	if err2 != nil {
		return err2
	}
	for _, osArch := range osArchs {
		if binFile, err := buildFor(ctx, commit, release, osArch[0], osArch[1], srcMainPath, execName); err != nil {
			return err
		} else {
			copyMaybeToLocalBin(binFile, osArch[0], osArch[1])
			zFile, err := createArchive(binFile, osArch[0])
			if err != nil {
				return err
			}
			fmt.Printf("deploy %s...", osArch[0])
			theKey := fmt.Sprintf("%s_%s", osArch[0], osArch[1])
			if err3 := svc.Upload(ctx, zFile, theKey); err3 != nil {
				return err3
			}
		}
	}
	return nil
}

func buildFor(ctx context.Context, commit string, release string, goos string, goarch string, srcMainPath string, execName string) (newfs.File, *cmd.XbeeError) {
	fmt.Printf("building %s for arch %s...", goos, goarch)
	binFile := localBin(goos, goarch, execName)
	binFile.Dir().EnsureExists()
	ldflagsRelease := ""
	if release != "" {
		ldflagsRelease = fmt.Sprintf(" -X 'github.com/iodasolutions/xbee-common/util.GitRelease=%s'", release)
	}
	//go build -ldflags "-X 'main.maVariable=$(echo $ENV_VAR)'"
	formattedTime := time.Now().Format("2006-01-02 15:04:05")
	ldflags := fmt.Sprintf("-X 'github.com/iodasolutions/xbee-common/util.BuildTime=%s' -X 'github.com/iodasolutions/xbee-common/util.GitCommit=%s'%s", formattedTime, commit, ldflagsRelease)
	ldflags += fmt.Sprintf(" -X 'github.com/iodasolutions/xbee-common/indus.n1=%s' -X 'github.com/iodasolutions/xbee-common/indus.s1=%s'", os.Getenv("XBEE_ACCESS_ID"), os.Getenv("XBEE_ACCESS_SECRET"))
	aCmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-gcflags", "all=-N -l", "-o", binFile.String(), fmt.Sprintf("%s/%s", newfs.CWD(), srcMainPath))
	aCmd.Env = environmentFor(goos, goarch)
	aCmd.Stderr = os.Stderr
	aCmd.Stdout = os.Stdout
	err := aCmd.Run()
	if err != nil {
		return "", cmd.Error("command %s %v failed: %v", aCmd.Path, strings.Join(aCmd.Args, " "), err)
	}
	fmt.Println("OK")
	return binFile, nil
}

func localBin(goos string, goarch string, execName string) newfs.File {
	if goos == "windows" {
		return localDir(publicBucket, "windows", goarch).ChildFile(execName + ".exe")
	} else {
		return localDir(publicBucket, goos, goarch).ChildFile(execName)
	}
}

func environmentFor(goos string, goarch string) []string {
	env := os.Environ()
	env = append(env, "CGO_ENABLED=0")
	env = append(env, fmt.Sprintf("GOOS=%s", goos))
	env = append(env, fmt.Sprintf("GOARCH=%s", goarch))
	return env
}
func localDir(bucket string, goos string, goarch string) newfs.Folder {
	return newfs.XbeeIntern().CacheArtefacts().ChildFolder(fmt.Sprintf("s3.eu-west-3.amazonaws.com/%s/%s_%s", bucket, goos, goarch))
}
