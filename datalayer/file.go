package datalayer

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type FileRepository interface {
	GetSignedURL(key string) (string, error)
	BuildImageURL(key string) string
}

type s3FileRepository struct {
	Service       *s3.S3
	UrlExpiration time.Duration
}

var (
	S3Bucket            = os.Getenv("S3_BUCKET")
	SignedURLExpiration = os.Getenv("SIGNED_URL_EXPIRATION")
)

func NewFileRepository(sess *session.Session) FileRepository {
	expiration, err := strconv.Atoi(SignedURLExpiration)
	if err != nil {
		expiration = 300
	}

	return &s3FileRepository{
		Service:       s3.New(sess),
		UrlExpiration: time.Duration(expiration) * time.Second,
	}
}

func (r *s3FileRepository) GetSignedURL(key string) (string, error) {
	req, _ := r.Service.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(S3Bucket),
		Key:    aws.String(key),
	})

	return req.Presign(r.UrlExpiration)
}

func (r *s3FileRepository) BuildImageURL(key string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", S3Bucket, key)
}
