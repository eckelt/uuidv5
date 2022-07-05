package backwards

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/eckelt/uuidv5/namespace"
	"github.com/google/uuid"
)

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

func Mandant(ctx context.Context, mId string) (string, error) {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	client := s3.NewFromConfig(cfg)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName("dev")),
		MaxKeys:   10000,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return "", err
	}

	namespaces, err := namespace.GetAllNamespaces(ctx)
	if err != nil {
		return "", err
	}

	for _, value := range output.CommonPrefixes {
		stripedPrefix := strings.Replace(aws.ToString(value.Prefix), "/", "", 1)
		if uuid.NewSHA1(namespaces["mandanten"], []byte(stripedPrefix)).String() == mId {
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

func idsFromS3(ctx context.Context, mandantId, typ string) ([]string, error) {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName("dev")),
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

func Rainbow(ctx context.Context, mandantId, typ string) (map[string]string, error) {

	namespaces, err := namespace.GetAllNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	ids, err := idsFromS3(ctx, mandantId, typ)
	if err != nil {
		return nil, err
	}

	return rainbow(mandantId, ids, namespaces[typ]), nil
}
