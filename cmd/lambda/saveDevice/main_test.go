package main

import (
	"context"
	"device-registry/pkg/model"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type mockSvc struct {
	saveErr error
}

func (m *mockSvc) GetDeviceByID(_ context.Context, _ string) (*model.Device, error) { return nil, nil }
func (m *mockSvc) SaveDevice(_ context.Context, _ *model.Device) error              { return m.saveErr }
func (m *mockSvc) DeleteDevice(_ context.Context, _ string) error                   { return nil }
func (m *mockSvc) GetAllDevices(_ context.Context, _ int, _ string) (*model.PageResult, error) {
	return &model.PageResult{}, nil
}
func (m *mockSvc) UpdateDevice(_ context.Context, _ *model.Device) error { return nil }

func setup(t *testing.T, m *mockSvc) {
	t.Helper()
	orig := svc
	t.Cleanup(func() { svc = orig })
	svc = m
}

func TestHandle_201_OnSuccess(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"deviceId":"id1","name":"router","type":"network"}`,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestHandle_400_InvalidJSON(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{Body: `not-json`})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandle_400_ValidationError(t *testing.T) {
	setup(t, &mockSvc{saveErr: &model.ValidationError{Message: "deviceId required"}})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{Body: `{}`})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandle_409_Conflict(t *testing.T) {
	setup(t, &mockSvc{saveErr: &model.ConflictError{ID: "id1"}})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"deviceId":"id1","name":"router","type":"network"}`,
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}