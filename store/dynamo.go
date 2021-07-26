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
	VoteConnectionIDs    []string
	ControlConnectionIDs []string
}

func (d *DynamoDB) loadDynamo(id string) (*dynamoItem, error) {
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
		dynamoKey: &dynamoKey{},
		Session:   &Session{},
	}
	err = attributevalue.UnmarshalMap(res.Item, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (d *DynamoDB) Load(id string) (*Session, error) {
	item, err := d.loadDynamo(id)
	if err != nil {
		return nil, err
	}

	return item.Session, nil
}

func (d *DynamoDB) Save(id string, item *domain.Session, version ...uuid.UUID) error {
	if len(version) > 1 {
		return ErrVersionMismatch
	}

	key, err := attributevalue.MarshalMap(newDynamoKey(id))
	if err != nil {
		return err
	}

	expr, err := expression.NewBuilder().
		WithUpdate(expression.
			Set(expression.Name("Data"), expression.Value(item)).
			Set(expression.Name("Version"), expression.Value(WithNewVersion(item).Version))).
		WithCondition(d.versionCheck(version...)).
		Build()
	if err != nil {
		return err
	}

	req := &dynamodb.UpdateItemInput{
		TableName:                 d.table,
		Key:                       key,
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = d.client.UpdateItem(context.TODO(), req)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return ErrVersionMismatch
		}

		return err
	}

	return nil
}

func (d *DynamoDB) AddVoteConnection(id string, connectionID string) error {
	return d.addConnection(id, "VoteConnectionIDs", connectionID)
}

func (d *DynamoDB) AddControlConnection(id string, connectionID string) error {
	return d.addConnection(id, "ControlConnectionIDs", connectionID)
}

func (d *DynamoDB) addConnection(id string, name, connectionID string) error {
	key, err := attributevalue.MarshalMap(newDynamoKey(id))
	if err != nil {
		return err
	}

	expr, err := expression.NewBuilder().
		WithUpdate(expression.
			Add(
				expression.Name(name),
				expression.Value(types.AttributeValueMemberSS{
					Value: []string{connectionID},
				}))).
		WithCondition(expression.AttributeExists(expression.Name("SessionID"))).
		Build()
	if err != nil {
		return err
	}

	req := &dynamodb.UpdateItemInput{
		TableName:                 d.table,
		Key:                       key,
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = d.client.UpdateItem(context.TODO(), req)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return ErrNotExists
		}

		return err
	}

	return nil
}

func (d *DynamoDB) GetVoteConnections(id string) ([]string, error) {
	item, err := d.loadDynamo(id)
	if err != nil {
		return nil, err
	}

	return item.VoteConnectionIDs, nil
}

func (d *DynamoDB) GetControlConnections(id string) ([]string, error) {
	item, err := d.loadDynamo(id)
	if err != nil {
		return nil, err
	}

	return item.ControlConnectionIDs, nil
}

func (d *DynamoDB) versionCheck(version ...uuid.UUID) expression.ConditionBuilder {
	if len(version) == 0 {
		return expression.AttributeNotExists(expression.Name("SessionID"))
	}

	return expression.Equal(
		expression.Name("Version"),
		expression.Value(version[0]))
}
