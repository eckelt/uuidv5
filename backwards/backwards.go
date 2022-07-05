package backwards

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func bucketName(stage string) string {
	return "immosolve-" + stage + "-classic-import-data"
}

func filename(key string) string {
	arr := strings.Split(key, "/")
	return arr[len(arr)-1]
}

var s3mapping = map[string]string{
	"personen":     "Person",
	"objekte":      "ImmoObject",
	"multimedia":   "Multimedia",
	"unternehmen":  "Firm",
	"aktivitaeten": "Activity",
}

func mandant(ctx context.Context, mId string, namespace uuid.UUID, client *s3.Client) (string, error) {

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName("dev")),
		MaxKeys:   10000,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return "", err
	}

	for _, value := range output.CommonPrefixes {
		stripedPrefix := strings.Replace(aws.ToString(value.Prefix), "/", "", 1)
		if uuid.NewSHA1(namespace, []byte(stripedPrefix)).String() == mId {
			return stripedPrefix, nil
		}
	}

	return "", errors.New("Mandant not found")
}

func Find(ctx context.Context, mId, ns string, namespaces map[string]uuid.UUID) (string, map[string]string, error) {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", nil, err
	}

	client := s3.NewFromConfig(cfg)

	mandantId, err := mandant(ctx, mId, namespaces["mandanten"], client)
	if err != nil {
		return "", nil, err
	}

	if ns == "mandanten" {
		return mandantId, nil, nil
	}
	
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName("dev")),
		Prefix:    aws.String(mandantId + "/" + s3mapping[ns] + "/"),
		MaxKeys:   1000,
		Delimiter: aws.String("/"),
	})

	uuids := map[string]string{}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", nil, err
		}
		for _, object := range page.Contents {
			fileArr := strings.Split(filename(aws.ToString(object.Key)), ".")
			file := strings.Join(fileArr[0:len(fileArr)-1], ".")
			// log.Printf("%s %s %s", prefix, file, uuid.NewSHA1(namespace, []byte(prefix+file)).String())
			uuids[uuid.NewSHA1(namespaces[ns], []byte(mandantId+file)).String()] = file
		}
	}

	return mandantId, uuids, nil
}
