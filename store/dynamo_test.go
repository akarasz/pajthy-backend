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

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	s := store.NewDynamoDB(&dynamoConfig, "testPajthy")
	suite.Run(t, &Suite{Subject: s})
}

func TestGetVoteConnections(t *testing.T) {
	s := store.NewDynamoDB(&dynamoConfig, "testPajthy")

	// get connections for a non-existing session should return error
	_, err := s.GetVoteConnections("getVoteConnections")
	assert.Equal(t, store.ErrNotExists, err)

	require.NoError(t, s.Save("getVoteConnections", domain.NewSession()))
	require.NoError(t, s.AddVoteConnection("getVoteConnections", "testID1"))
	require.NoError(t, s.AddVoteConnection("getVoteConnections", "testID2"))

	got, err := s.GetVoteConnections("getVoteConnections")
	assert.NoError(t, err)
	assert.Contains(t, got, "testID1")
	assert.Contains(t, got, "testID2")
}

func TestGetControlConnections(t *testing.T) {
	s := store.NewDynamoDB(&dynamoConfig, "testPajthy")

	// get connections for a non-existing session should return error
	_, err := s.GetControlConnections("getControlConnections")
	assert.Equal(t, store.ErrNotExists, err)

	require.NoError(t, s.Save("getControlConnections", domain.NewSession()))
	require.NoError(t, s.AddControlConnection("getControlConnections", "testID1"))
	require.NoError(t, s.AddControlConnection("getControlConnections", "testID2"))

	got, err := s.GetControlConnections("getControlConnections")
	assert.NoError(t, err)
	assert.Contains(t, got, "testID1")
	assert.Contains(t, got, "testID2")
}

func TestAddVoteConnection(t *testing.T) {
	s := store.NewDynamoDB(&dynamoConfig, "testPajthy")

	// add a connection to a non-existing session should return error
	err := s.AddVoteConnection("addVoteConnection", "testID")
	assert.Equal(t, store.ErrNotExists, err)

	require.NoError(t, s.Save("addVoteConnection", domain.NewSession()))

	// adding a connection to an existing session should be ok
	err = s.AddVoteConnection("addVoteConnection", "testID")
	assert.NoError(t, err)

	// adding another one should be still ok
	err = s.AddVoteConnection("addVoteConnection", "testID2")
	assert.NoError(t, err)

	got, err := s.GetVoteConnections("addVoteConnection")
	require.NoError(t, err)

	// should contain the added testID
	assert.Contains(t, got, "testID")
	assert.Contains(t, got, "testID2")
}

func TestAddControlConnection(t *testing.T) {
	s := store.NewDynamoDB(&dynamoConfig, "testPajthy")

	// add a connection to a non-existing session should return error
	err := s.AddControlConnection("addControlConnection", "testID")
	assert.Equal(t, store.ErrNotExists, err)

	require.NoError(t, s.Save("addControlConnection", domain.NewSession()))

	// adding a connection to an existing session should be ok
	err = s.AddControlConnection("addControlConnection", "testID")
	assert.NoError(t, err)

	// adding another one should be still ok
	err = s.AddControlConnection("addControlConnection", "testID2")
	assert.NoError(t, err)

	got, err := s.GetControlConnections("addControlConnection")
	require.NoError(t, err)

	// should contain the added testID
	assert.Contains(t, got, "testID")
	assert.Contains(t, got, "testID2")
}
