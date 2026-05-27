package main

import (
	"context"
	"device-registry/pkg/model"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type mockSvc struct {
	deleteErr error
}

func (m *mockSvc) GetDeviceByID(_ context.Context, _ string) (*model.Device, error) { return nil, nil }
func (m *mockSvc) SaveDevice(_ context.Context, _ *model.Device) error              { return nil }
func (m *mockSvc) DeleteDevice(_ context.Context, _ string) error                   { return m.deleteErr }
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

func req(id string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{PathParameters: map[string]string{"id": id}}
}

func TestHandle_200_Deleted(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), req("id1"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandle_404_NotFound(t *testing.T) {
	setup(t, &mockSvc{deleteErr: &model.NotFoundError{ID: "missing"}})
	resp, _ := handle(context.Background(), req("missing"))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandle_400_MissingID(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}