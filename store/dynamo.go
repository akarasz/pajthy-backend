package store

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

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

func (d *DynamoDB) Create(id string, created *domain.Session) error {
	item, err := attributevalue.MarshalMap(newDynamoItem(id, WithNewVersion(created)))
	if err != nil {
		return err
	}

	notExists := expression.AttributeNotExists(expression.Name("SessionID"))
	expr, err := expression.NewBuilder().
		WithCondition(notExists).
		Build()
	if err != nil {
		return err
	}

	req := &dynamodb.PutItemInput{
		TableName:                d.table,
		Item:                     item,
		ConditionExpression:      expr.Condition(),
		ExpressionAttributeNames: expr.Names(),
	}

	_, err = d.client.PutItem(context.TODO(), req)
	if err != nil {
		return err
	}

	return nil
}

func (d *DynamoDB) Update(id string, updated *Session) error {
	item, err := attributevalue.MarshalMap(newDynamoItem(id, WithNewVersion(updated.Data)))
	if err != nil {
		return err
	}

	matchingId := expression.Equal(expression.Name("Version"), expression.Value(updated.Version))
	expr, err := expression.NewBuilder().
		WithCondition(matchingId).
		Build()
	if err != nil {
		return err
	}

	req := &dynamodb.PutItemInput{
		TableName:                 d.table,
		Item:                      item,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = d.client.PutItem(context.TODO(), req)
	if err != nil {
		return err
	}

	return nil
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
