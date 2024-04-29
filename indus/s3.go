package indus

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/log2"
	"github.com/iodasolutions/xbee-common/newfs"
	"log"
	"os"
	"strings"
)

var n1 string
var s1 string

type S3 struct {
	s3     *s3.Client
	bucket string
}

func AdminClient(bucket string) (*S3, *cmd.XbeeError) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, cmd.Error("cannot create s3 client: %v", err)
	}
	return &S3{s3.NewFromConfig(cfg), bucket}, nil
}

func ReadClient() *S3 {
	cfg := aws.Config{
		Region:      "eu-west-3",
		Credentials: credentials.NewStaticCredentialsProvider(n1, s1, ""),
	}
	return &S3{s3.NewFromConfig(cfg), PrivateBucket}
}

func (svc *S3) HasEntry(ctx context.Context, key string) (bool, *cmd.XbeeError) {
	_, err := svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &svc.bucket,
		Key:    &key,
	})
	if err != nil {
		message := err.Error()
		if strings.Contains(message, "StatusCode: 404") {
			return false, nil
		} else {
			return false, cmd.Error("Unknown S3 Error: %v", err)
		}
	}
	return true, nil
}

func (svc *S3) Upload(ctx context.Context, f newfs.File, key string) *cmd.XbeeError {
	file, err2 := os.Open(f.String())
	if err2 != nil {
		return cmd.Error("cannot open file %s: %v", f, err2)
	}
	defer file.Close()
	_, err2 = svc.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &svc.bucket,
		Key:    &key,
		Body:   file,
	})
	if err2 != nil {
		return cmd.Error("cannot upload file %s to s3: %v", f, err2)
	}
	fmt.Printf("Upload OK (%s)\n", f)
	return nil
}

func (svc *S3) Download(ctx context.Context, key string, targetFile newfs.File) *cmd.XbeeError {
	f, err := targetFile.OpenFileForCreation()
	if err != nil {
		return err
	}
	defer f.Close()
	downloader := manager.NewDownloader(svc.s3)
	// Télécharger l'objet
	_, err2 := downloader.Download(ctx, f, &s3.GetObjectInput{
		Bucket: &PrivateBucket,
		Key:    &key,
	})
	if err2 != nil {
		log.Fatalf("Unable to download item %s to %s, %v", key, targetFile, err)
	}
	log2.Infof("downloaded %s to %s from s3", key, targetFile)
	return nil
}
