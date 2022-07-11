package awsExports

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/google/uuid"
)

type Exports struct {
	client *cloudformation.Client
}

func NewFromAwsConfig(cfg aws.Config) *Exports {
	return &Exports{
		client: cloudformation.NewFromConfig(cfg),
	}
}

func (e *Exports) getAllExports(ctx context.Context) ([]types.Export, error) {
	output, err := e.client.ListExports(ctx, &cloudformation.ListExportsInput{})
	if err != nil {
		return nil, err
	}

	return output.Exports, nil
}

func (e *Exports) Namespaces(ctx context.Context) (map[string]uuid.UUID, error) {
	exports, err := e.getAllExports(ctx)
	if err != nil {
		return nil, err
	}
	namespaces := map[string]uuid.UUID{}
	for _, object := range exports {
		key := aws.ToString(object.Name)
		if strings.HasSuffix(key, "-namespace") {
			uuid, err := uuid.Parse(aws.ToString(object.Value))
			if err != nil {
				return nil, err
			}
			newKey := strings.Replace(key, "-namespace", "", 1)
			namespaces[newKey] = uuid
		}
	}
	return namespaces, nil
}

func (e *Exports) Stage(ctx context.Context) (string, error) {
	exports, err := e.getAllExports(ctx)
	if err != nil {
		return "", err
	}
	for _, export := range exports {
		if aws.ToString(export.Name) == "stage" {
			return aws.ToString(export.Value), nil
		}
	}
	return "", errors.New("stage not found")
}
