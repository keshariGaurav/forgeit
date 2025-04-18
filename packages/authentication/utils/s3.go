package utils

import (
	"authentication/configs"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func InitS3() (*s3.Client, string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg)
	bucketName := configs.EnvAWSBucketName()
	return client, bucketName
}

func UploadToS3(client *s3.Client, bucketName string, file io.Reader, originalFilename string) (string, error) {
	key := fmt.Sprintf("resumes/%s-%s", uuid.New().String(), originalFilename)

	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", err
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key)
	return url, nil
}
