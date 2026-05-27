package repository

import (
	"context"
	"device-registry/pkg/model"
	"encoding/base64"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type DeviceRepository struct {
	svc       dynamoClient
	tableName string
}

func NewDeviceRepository() (*DeviceRepository, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		tableName = "Devices"
	}
	return &DeviceRepository{
		svc:       dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

func newWithClient(client dynamoClient, tableName string) *DeviceRepository {
	return &DeviceRepository{svc: client, tableName: tableName}
}

func (r *DeviceRepository) GetDeviceByID(ctx context.Context, deviceID string) (*model.Device, error) {
	result, err := r.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"deviceId": &types.AttributeValueMemberS{Value: deviceID},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Item) == 0 {
		return nil, nil
	}
	var device model.Device
	err = attributevalue.UnmarshalMap(result.Item, &device)
	return &device, err
}

func (r *DeviceRepository) SaveDevice(ctx context.Context, device *model.Device) error {
	_, err := r.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		ConditionExpression: aws.String("attribute_not_exists(deviceId)"),
		Item: map[string]types.AttributeValue{
			"deviceId": &types.AttributeValueMemberS{Value: device.DeviceID},
			"name":     &types.AttributeValueMemberS{Value: device.Name},
			"type":     &types.AttributeValueMemberS{Value: device.Type},
		},
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &model.ConflictError{ID: device.DeviceID}
		}
		return err
	}
	return nil
}

func (r *DeviceRepository) DeleteDevice(ctx context.Context, deviceID string) error {
	_, err := r.svc.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           &r.tableName,
		ConditionExpression: aws.String("attribute_exists(deviceId)"),
		Key: map[string]types.AttributeValue{
			"deviceId": &types.AttributeValueMemberS{Value: deviceID},
		},
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &model.NotFoundError{ID: deviceID}
		}
		return err
	}
	return nil
}

func (r *DeviceRepository) GetAllDevices(ctx context.Context, limit int, nextToken string) (*model.PageResult, error) {
	input := &dynamodb.ScanInput{
		TableName: &r.tableName,
	}
	if limit > 0 {
		l := int32(limit)
		input.Limit = &l
	}
	if nextToken != "" {
		raw, err := base64.StdEncoding.DecodeString(nextToken)
		if err != nil {
			return nil, &model.ValidationError{Message: "invalid nextToken"}
		}
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"deviceId": &types.AttributeValueMemberS{Value: string(raw)},
		}
	}

	result, err := r.svc.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	devices := make([]model.Device, 0, len(result.Items))
	if err = attributevalue.UnmarshalListOfMaps(result.Items, &devices); err != nil {
		return nil, err
	}

	pageResult := &model.PageResult{Devices: devices}
	if v, ok := result.LastEvaluatedKey["deviceId"].(*types.AttributeValueMemberS); ok {
		pageResult.NextToken = base64.StdEncoding.EncodeToString([]byte(v.Value))
	}
	return pageResult, nil
}

func (r *DeviceRepository) UpdateDevice(ctx context.Context, device *model.Device) error {
	_, err := r.svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           &r.tableName,
		ConditionExpression: aws.String("attribute_exists(deviceId)"),
		Key: map[string]types.AttributeValue{
			"deviceId": &types.AttributeValueMemberS{Value: device.DeviceID},
		},
		UpdateExpression: aws.String("set #n = :n, #t = :t"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
			"#t": "type",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":n": &types.AttributeValueMemberS{Value: device.Name},
			":t": &types.AttributeValueMemberS{Value: device.Type},
		},
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &model.NotFoundError{ID: device.DeviceID}
		}
		return err
	}
	return nil
}