package store_test

import (
	"testing"
	// "github.com/stretchr/testify/suite"
	// "github.com/akarasz/pajthy-backend/store"
)

func TestDynamoDB(t *testing.T) {
	// TODO
	// customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
	// 	if service == dynamodb.ServiceID {
	// 		return aws.Endpoint{
	// 			PartitionID:   "aws",
	// 			URL:           "http://localhost:8000",
	// 			SigningRegion: region,
	// 		}, nil
	// 	}
	// 	return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	// })

	// c, err := config.LoadDefaultConfig(context.TODO() /*, config.WithEndpointResolver(customResolver)*/)
	// if err != nil {
	// return nil, err
	// }

	// s := store.NewDynamoDB(nil, "")
	// suite.Run(t, &Suite{Subject: s})
}
