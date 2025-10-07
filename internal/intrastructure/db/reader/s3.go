package reader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Reader struct {
	cfg aws.Config
}

func NewS3Reader(region, endpoint, accessKey, secretKey string, forcePathStyle bool) (*S3Reader, error) {
	optFns := []func(*awsCfg.LoadOptions) error{}

	// Region
	if region == "" {
		region = "us-east-1"
	}
	optFns = append(optFns, awsCfg.WithRegion(region))

	// Endpoint custom (MinIO, etc.)
	if endpoint != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, _ ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
				SigningRegion:     region,
			}, nil
		})
		optFns = append(optFns, awsCfg.WithEndpointResolverWithOptions(resolver))
	}

	if accessKey != "" && secretKey != "" {
		optFns = append(optFns,
			awsCfg.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			),
		)
	}

	cfg, err := awsCfg.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, err
	}

	return &S3Reader{cfg: cfg}, nil
}

func (s *S3Reader) Open(s3url string) (io.ReadCloser, error) {
	if !strings.HasPrefix(s3url, "s3://") {
		return nil, fmt.Errorf("ruta no es s3://")
	}
	u, err := url.Parse(s3url)
	if err != nil {
		return nil, err
	}
	bucket := u.Host
	key := strings.TrimPrefix(u.Path, "/")

	client := s3.NewFromConfig(s.cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	buf := manager.NewWriteAtBuffer(nil)
	d := manager.NewDownloader(client)
	_, err = d.Download(context.Background(), buf, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
