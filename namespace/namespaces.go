package namespace

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/google/uuid"
)

func collect(exports []types.Export) map[string]uuid.UUID {
	namespaces := map[string]uuid.UUID{}
	for _, object := range exports {
		key := aws.ToString(object.Name)
		if strings.HasSuffix(key, "-namespace") {
			uuid, err := uuid.Parse(aws.ToString(object.Value))
			if err != nil {
				log.Println(err)
			}
			newKey := strings.Replace(key, "-namespace", "", 1)
			namespaces[newKey] = uuid
		}
	}
	return namespaces
}

func GetAllNamespaces(ctx context.Context) (map[string]uuid.UUID, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := cloudformation.NewFromConfig(cfg)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListExports(ctx, &cloudformation.ListExportsInput{})
	if err != nil {
		return nil, err
	}

	return collect(output.Exports), nil
}
