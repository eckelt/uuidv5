package namespace

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
)

func TestConvert(t *testing.T) {
	// Given
	randUuid1 := uuid.New()
	randUuid2 := uuid.New()
	exports := []types.Export{
		{Name: aws.String("objekt-namespace"), Value: aws.String(randUuid1.String())},
		{Name: aws.String("person-namespace"), Value: aws.String(randUuid2.String())},
		{Name: aws.String("hans-wurst"), Value: aws.String("asdf")},
	}

	// When
	actual := collect(exports)

	// Then
	expected := map[string]uuid.UUID{
		"person": randUuid2,
		"objekt": randUuid1,
	}
	
	compareMaps(actual, expected, t)
}


func compareMaps(actual, expected map[string]uuid.UUID, t *testing.T) {
	for key, uuid := range expected {
		actualUuid, exists := actual[key]
		if !exists {
			fmt.Printf("Expected key %s to be found\n", key)
			t.Fail()
		}
		if actualUuid != uuid {
			fmt.Printf("Expected key %s to be %s\n", key, uuid)
			t.Fail()
		}

	}
}