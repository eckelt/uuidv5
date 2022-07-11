package s3Data

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Data struct {
	stage      string
	client     *s3.Client
	namespaces map[string]uuid.UUID
}

func NewFromAwsConfig(cfg aws.Config, stage string, namespaces map[string]uuid.UUID) *S3Data {
	return &S3Data{
		stage:      stage,
		client:     s3.NewFromConfig(cfg),
		namespaces: namespaces,
	}
}

func bucketName(stage string) string {
	return "immosolve-" + stage + "-classic-import-data"
}

func filename(key string) string {
	file := strings.Split(key, "/")
	fileArr := strings.Split(file[len(file)-1], ".")
	return strings.Join(fileArr[0:len(fileArr)-1], ".")
}

var s3mapping = map[string]string{
	"personen":     "Person",
	"objekte":      "ImmoObject",
	"multimedia":   "Multimedia",
	"unternehmen":  "Firm",
	"aktivitaeten": "Activity",
}

func (s *S3Data) Mandant(ctx context.Context, mId string) (string, error) {
	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName(s.stage)),
		MaxKeys:   10000,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return "", err
	}

	for _, value := range output.CommonPrefixes {
		stripedPrefix := strings.Replace(aws.ToString(value.Prefix), "/", "", 1)
		if uuid.NewSHA1(s.namespaces["mandanten"], []byte(stripedPrefix)).String() == mId {
			return stripedPrefix, nil
		}
	}

	return "", errors.New("Mandant not found")
}

func rainbow(mandantId string, ids []string, namespace uuid.UUID) map[string]string {
	uuids := map[string]string{}
	for _, id := range ids {
		uuids[uuid.NewSHA1(namespace, []byte(mandantId+id)).String()] = id
	}
	return uuids
}

func (s *S3Data) idsFromS3(ctx context.Context, mandantId, typ string) ([]string, error) {

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName(s.stage)),
		Prefix:    aws.String(mandantId + "/" + s3mapping[typ] + "/"),
		MaxKeys:   1000,
		Delimiter: aws.String("/"),
	})

	ids := []string{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, object := range page.Contents {
			id := filename(aws.ToString(object.Key))
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (s *S3Data) Rainbow(ctx context.Context, mandantId, typ string) (map[string]string, error) {

	ids, err := s.idsFromS3(ctx, mandantId, typ)
	if err != nil {
		return nil, err
	}

	return rainbow(mandantId, ids, s.namespaces[typ]), nil
}
