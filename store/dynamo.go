package store

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/akarasz/pajthy-backend/domain"
)

type DynamoDB struct {
	client *dynamodb.Client
	table  *string
}

func NewDynamoDB(c *aws.Config, table string) *DynamoDB {
	client := dynamodb.NewFromConfig(*c)
	return &DynamoDB{
		client: client,
		table:  aws.String(table),
	}
}

type dynamoKey struct {
	SessionID string
}

func newDynamoKey(id string) *dynamoKey {
	return &dynamoKey{
		SessionID: id,
	}
}

type dynamoItem struct {
	*dynamoKey
	*Session
}

func newDynamoItem(id string, s *Session) *dynamoItem {
	return &dynamoItem{
		newDynamoKey(id),
		s,
	}
}

func (d *DynamoDB) Load(id string) (*Session, error) {
	key, err := attributevalue.MarshalMap(newDynamoKey(id))
	if err != nil {
		return nil, err
	}

	req := &dynamodb.GetItemInput{
		TableName: d.table,
		Key:       key,
	}

	res, err := d.client.GetItem(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	if len(res.Item) == 0 {
		return nil, ErrNotExists
	}

	item := dynamoItem{
		&dynamoKey{},
		&Session{},
	}
	err = attributevalue.UnmarshalMap(res.Item, &item)
	if err != nil {
		return nil, err
	}

	return item.Session, nil
}

func (d *DynamoDB) Save(id string, item *domain.Session, version ...uuid.UUID) error {
	if len(version) > 1 {
		return ErrVersionMismatch
	}

	data, err := attributevalue.MarshalMap(newDynamoItem(id, WithNewVersion(item)))
	if err != nil {
		return err
	}

	expr, err := expression.NewBuilder().
		WithCondition(dynamoCondition(version...)).
		Build()
	if err != nil {
		return err
	}

	req := &dynamodb.PutItemInput{
		TableName:                 d.table,
		Item:                      data,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = d.client.PutItem(context.TODO(), req)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return ErrVersionMismatch
		}

		return err
	}

	return nil
}

func dynamoCondition(version ...uuid.UUID) expression.ConditionBuilder {
	if len(version) == 0 {
		return expression.AttributeNotExists(expression.Name("SessionID"))
	}

	return expression.Equal(
		expression.Name("Version"),
		expression.Value(version[0]))
}
