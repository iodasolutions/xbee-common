package indus

import (
	"context"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"os"
	"os/exec"
	"strings"
)

const publicBucket = "xbee.repository.public"

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
	for _, osArch := range osArchs {
		if targetBin, err := buildFor(ctx, osArch[0], osArch[1], srcMainPath, execName); err != nil {
			return err
		} else {
			copyMaybeToLocalBin(targetBin, osArch[0], osArch[1])
		}
	}
	return nil
}

func BuildAndDeploy(ctx context.Context, srcMainPath string, execName string) *cmd.XbeeError {
	for _, osArch := range osArchs {
		if binFile, err := buildFor(ctx, osArch[0], osArch[1], srcMainPath, execName); err != nil {
			return err
		} else {
			copyMaybeToLocalBin(binFile, osArch[0], osArch[1])
			var zFile newfs.File
			if osArch[0] == "windows" {
				zFile = binFile.DoZip()
				fmt.Println("zipped OK")
			} else {
				zFile = binFile.DoTargGz()
				fmt.Println("tar.gz OK")
			}

			fmt.Printf("deploy %s...", osArch[0])
			u := Unit{
				Source: zFile,
				Bucket: publicBucket,
				Key:    fmt.Sprintf("%s_%s", osArch[0], osArch[1]),
			}
			if err := u.UploadToS3(ctx); err != nil {
				return err
			}

		}
	}
	return nil
}

func buildFor(ctx context.Context, goos string, goarch string, srcMainPath string, execName string) (newfs.File, *cmd.XbeeError) {
	fmt.Printf("building %s for arch %s...", goos, goarch)
	binFile := localBin(goos, goarch, execName)
	binFile.Dir().EnsureEmpty()
	aCmd := exec.CommandContext(ctx, "go", "build", "-gcflags", "all=-N -l", "-o", binFile.String(), fmt.Sprintf("%s/%s", newfs.CWD(), srcMainPath))
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