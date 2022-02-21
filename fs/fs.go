package fs

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	ctx    context.Context
	Cfg    Config
	Client *minio.Client
)

type Config struct {
	EndPoint           string
	Port               int
	Region             string
	BucketName         string
	AccessKey          string
	SecretKey          string
	IsAutoCreateBucket bool
}

func Configure(c Config) error {
	var err error
	Client, err = minio.New(c.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return err
	}
	Cfg = c
	ctx = context.Background()
	if c.IsAutoCreateBucket {
		isExist, errBucketExist := Client.BucketExists(ctx, c.BucketName)
		if !isExist {
			err := Client.MakeBucket(ctx, c.BucketName, minio.MakeBucketOptions{Region: c.Region})
			if err != nil {
				return err
			}
		} else if errBucketExist != nil {
			return errBucketExist
		}
	}
	return nil
}

func GetPath(fileName string, path ...string) string {
	if Cfg.EndPoint != "s3.amazonaws.com" {
		return fileName
	}
	res := ""
	for _, p := range path {
		res += p + "/"
	}
	res += fileName

	return res
}

func GetUrl(fileName string, path ...string) string {
	if Cfg.EndPoint != "s3.amazonaws.com" {
		return "https://" + Cfg.BucketName + "." + Cfg.EndPoint + "/" + fileName
	}
	res := "https://" + Cfg.BucketName + ".s3." + Cfg.Region + ".amazonaws.com/"
	for _, p := range path {
		res += p + "/"
	}
	res += fileName

	return res
}

func Upload(fileName string, file *io.Reader, fileSize int64, opts ...minio.PutObjectOptions) (minio.UploadInfo, error) {
	opt := minio.PutObjectOptions{}
	opt.UserMetadata = map[string]string{"x-amz-acl": "public-read"}
	if len(opts) > 0 {
		opt = opts[0]
	}
	return Client.PutObject(ctx, Cfg.BucketName, fileName, *file, fileSize, opt)
}

func Delete(fileName string, opts ...minio.RemoveObjectOptions) error {
	opt := minio.RemoveObjectOptions{}
	opt.GovernanceBypass = true
	if len(opts) > 0 {
		opt = opts[0]
	}
	return Client.RemoveObject(ctx, Cfg.BucketName, fileName, opt)
}
