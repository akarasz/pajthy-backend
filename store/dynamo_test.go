package store_test

import (
	"context"
	"fmt"
	"testing"
	// "time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/akarasz/pajthy-backend/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDynamoDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping store/redis test")
	}

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForListeningPort("8000/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	ip, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "8000")
	require.NoError(t, err)

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

	c, err := config.LoadDefaultConfig(ctx, config.WithEndpointResolver(customResolver))
	require.NoError(t, err)

	// TODO create testPajthy table
	client := dynamodb.NewFromConfig(c)

	_, err = client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
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
	require.NoError(t, err)

	s := store.NewDynamoDB(&c, "testPajthy")
	suite.Run(t, &Suite{Subject: s})

	// time.Sleep(5 * time.Minute)
}
