package main

import (
	"context"
	"device-registry/pkg/model"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type mockSvc struct {
	result *model.PageResult
	err    error
}

func (m *mockSvc) GetDeviceByID(_ context.Context, _ string) (*model.Device, error) { return nil, nil }
func (m *mockSvc) SaveDevice(_ context.Context, _ *model.Device) error              { return nil }
func (m *mockSvc) DeleteDevice(_ context.Context, _ string) error                   { return nil }
func (m *mockSvc) GetAllDevices(_ context.Context, _ int, _ string) (*model.PageResult, error) {
	return m.result, m.err
}
func (m *mockSvc) UpdateDevice(_ context.Context, _ *model.Device) error { return nil }

func setup(t *testing.T, m *mockSvc) {
	t.Helper()
	orig := svc
	t.Cleanup(func() { svc = orig })
	svc = m
}

func TestHandle_200_ReturnsPageResult(t *testing.T) {
	setup(t, &mockSvc{result: &model.PageResult{
		Devices: []model.Device{{DeviceID: "id1", Name: "r", Type: "net"}},
	}})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var page model.PageResult
	if err := json.Unmarshal([]byte(resp.Body), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(page.Devices))
	}
}

func TestHandle_400_InvalidLimit(t *testing.T) {
	setup(t, &mockSvc{result: &model.PageResult{}})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"limit": "abc"},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandle_400_LimitOutOfRange(t *testing.T) {
	setup(t, &mockSvc{result: &model.PageResult{}})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"limit": "9999"},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandle_400_InvalidNextToken(t *testing.T) {
	setup(t, &mockSvc{err: &model.ValidationError{Message: "invalid nextToken"}})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"nextToken": "!!!"},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}