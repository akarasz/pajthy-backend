package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/suite"

	"github.com/akarasz/pajthy-backend/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var dynamoConfig aws.Config

func dynamoSetup() (teardown func(), err error) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForListeningPort("8000/tcp"),
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "8000")
	if err != nil {
		return nil, err
	}

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == dynamodb.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           fmt.Sprintf("http://%s:%s", ip, port.Port()),
				SigningRegion: region,
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	dynamoConfig, err = config.LoadDefaultConfig(ctx, config.WithEndpointResolver(customResolver))
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(dynamoConfig)

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("testPajthy"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("SessionID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("SessionID"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})
	if err != nil {
		return nil, err
	}

	return func() {
		container.Terminate(ctx)
	}, nil
}

func TestSuiteWithDynamo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping store/dynamo test")
	}

	s := store.NewDynamoDB(&dynamoConfig, "testPajthy")
	suite.Run(t, &Suite{Subject: s})
}

func TestAddConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping store/dynamo test")
	}

}
