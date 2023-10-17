package indus

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"os"
)

type Unit struct {
	Source newfs.File
	Bucket string
	Key    string
	Sha1   newfs.File
}

func (u *Unit) UploadToS3(ctx context.Context) *cmd.XbeeError {
	if err := uploadFile(ctx, u.Source, u.Bucket, u.Key); err != nil {
		return err
	}
	if u.Sha1 != "" {
		if err := uploadFile(ctx, u.Sha1, u.Bucket, u.Key); err != nil {
			return err
		}
	}
	return nil
}

func adminClient() (*s3.Client, *cmd.XbeeError) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, cmd.Error("cannot create s3 client: %v", err)
	}
	return s3.NewFromConfig(cfg), nil
}

func uploadFile(ctx context.Context, f newfs.File, bucket string, path string) *cmd.XbeeError {
	client, err := adminClient()
	if err != nil {
		return err
	}
	key := path + "/" + f.Base()
	file, err2 := os.Open(f.String())
	if err2 != nil {
		return cmd.Error("cannot open file %s: %v", f, err2)
	}
	defer file.Close()
	_, err2 = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	})
	if err2 != nil {
		return cmd.Error("cannot upload file %s to s3: %v", f, err2)
	}
	fmt.Printf("Upload OK (%s)\n", f)
	return nil
}
