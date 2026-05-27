package repository

import (
	"context"
	"device-registry/pkg/model"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type mockDynamo struct {
	getItemOut    *dynamodb.GetItemOutput
	getItemErr    error
	putItemErr    error
	deleteItemErr error
	scanOut       *dynamodb.ScanOutput
	scanErr       error
	updateItemErr error
}

func (m *mockDynamo) GetItem(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getItemOut, m.getItemErr
}
func (m *mockDynamo) PutItem(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, m.putItemErr
}
func (m *mockDynamo) DeleteItem(_ context.Context, _ *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, m.deleteItemErr
}
func (m *mockDynamo) Scan(_ context.Context, _ *dynamodb.ScanInput, _ ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return m.scanOut, m.scanErr
}
func (m *mockDynamo) UpdateItem(_ context.Context, _ *dynamodb.UpdateItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return &dynamodb.UpdateItemOutput{}, m.updateItemErr
}

func mustMarshalMap(v any) map[string]types.AttributeValue {
	m, err := attributevalue.MarshalMap(v)
	if err != nil {
		panic(err)
	}
	return m
}

func mustMarshalList(devices []model.Device) []map[string]types.AttributeValue {
	result := make([]map[string]types.AttributeValue, len(devices))
	for i, d := range devices {
		m, err := attributevalue.MarshalMap(d)
		if err != nil {
			panic(err)
		}
		result[i] = m
	}
	return result
}

var ctx = context.Background()

func TestGetDeviceByID_Found(t *testing.T) {
	repo := newWithClient(&mockDynamo{
		getItemOut: &dynamodb.GetItemOutput{
			Item: mustMarshalMap(model.Device{DeviceID: "id1", Name: "router", Type: "net"}),
		},
	}, "Devices")
	d, err := repo.GetDeviceByID(ctx, "id1")
	if err != nil || d == nil || d.Name != "router" {
		t.Fatalf("unexpected: %v %v", d, err)
	}
}

func TestGetDeviceByID_NotFound(t *testing.T) {
	repo := newWithClient(&mockDynamo{getItemOut: &dynamodb.GetItemOutput{}}, "Devices")
	d, err := repo.GetDeviceByID(ctx, "missing")
	if err != nil || d != nil {
		t.Fatalf("expected nil device, got %v %v", d, err)
	}
}

func TestGetDeviceByID_Error(t *testing.T) {
	repo := newWithClient(&mockDynamo{getItemErr: errors.New("dynamo error")}, "Devices")
	_, err := repo.GetDeviceByID(ctx, "id1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSaveDevice_OK(t *testing.T) {
	repo := newWithClient(&mockDynamo{}, "Devices")
	if err := repo.SaveDevice(ctx, &model.Device{DeviceID: "id1", Name: "r", Type: "net"}); err != nil {
		t.Fatal(err)
	}
}

func TestSaveDevice_Conflict(t *testing.T) {
	repo := newWithClient(&mockDynamo{
		putItemErr: &types.ConditionalCheckFailedException{},
	}, "Devices")
	err := repo.SaveDevice(ctx, &model.Device{DeviceID: "id1", Name: "r", Type: "net"})
	var ce *model.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("expected ConflictError, got %v", err)
	}
}

func TestDeleteDevice_OK(t *testing.T) {
	repo := newWithClient(&mockDynamo{}, "Devices")
	if err := repo.DeleteDevice(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteDevice_NotFound(t *testing.T) {
	repo := newWithClient(&mockDynamo{
		deleteItemErr: &types.ConditionalCheckFailedException{},
	}, "Devices")
	err := repo.DeleteDevice(ctx, "missing")
	var nfe *model.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestGetAllDevices_OK(t *testing.T) {
	items := mustMarshalList([]model.Device{
		{DeviceID: "id1", Name: "a", Type: "x"},
		{DeviceID: "id2", Name: "b", Type: "y"},
	})
	repo := newWithClient(&mockDynamo{
		scanOut: &dynamodb.ScanOutput{Items: items},
	}, "Devices")
	result, err := repo.GetAllDevices(ctx, 100, "")
	if err != nil || len(result.Devices) != 2 || result.NextToken != "" {
		t.Fatalf("unexpected: %v %v", result, err)
	}
}

func TestGetAllDevices_NextTokenRoundtrip(t *testing.T) {
	items := mustMarshalList([]model.Device{{DeviceID: "id1", Name: "a", Type: "x"}})
	repo := newWithClient(&mockDynamo{
		scanOut: &dynamodb.ScanOutput{
			Items: items,
			LastEvaluatedKey: map[string]types.AttributeValue{
				"deviceId": &types.AttributeValueMemberS{Value: "id1"},
			},
		},
	}, "Devices")
	result, err := repo.GetAllDevices(ctx, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	expected := base64.StdEncoding.EncodeToString([]byte("id1"))
	if result.NextToken != expected {
		t.Fatalf("expected nextToken %q, got %q", expected, result.NextToken)
	}
}

func TestGetAllDevices_InvalidNextToken(t *testing.T) {
	repo := newWithClient(&mockDynamo{}, "Devices")
	_, err := repo.GetAllDevices(ctx, 10, "not-valid-base64!!!")
	var ve *model.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestUpdateDevice_OK(t *testing.T) {
	repo := newWithClient(&mockDynamo{}, "Devices")
	if err := repo.UpdateDevice(ctx, &model.Device{DeviceID: "id1", Name: "new", Type: "sensor"}); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDevice_NotFound(t *testing.T) {
	repo := newWithClient(&mockDynamo{
		updateItemErr: &types.ConditionalCheckFailedException{},
	}, "Devices")
	err := repo.UpdateDevice(ctx, &model.Device{DeviceID: "missing", Name: "x", Type: "y"})
	var nfe *model.NotFoundError
	if !errors.As(err, &nfe) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}