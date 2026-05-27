package main

import (
	"context"
	"device-registry/pkg/model"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type mockSvc struct {
	updateErr error
}

func (m *mockSvc) GetDeviceByID(_ context.Context, _ string) (*model.Device, error) { return nil, nil }
func (m *mockSvc) SaveDevice(_ context.Context, _ *model.Device) error              { return nil }
func (m *mockSvc) DeleteDevice(_ context.Context, _ string) error                   { return nil }
func (m *mockSvc) GetAllDevices(_ context.Context, _ int, _ string) (*model.PageResult, error) {
	return &model.PageResult{}, nil
}
func (m *mockSvc) UpdateDevice(_ context.Context, _ *model.Device) error { return m.updateErr }

func setup(t *testing.T, m *mockSvc) {
	t.Helper()
	orig := svc
	t.Cleanup(func() { svc = orig })
	svc = m
}

func req(id, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"id": id},
		Body:           body,
	}
}

func TestHandle_200_Updated(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), req("id1", `{"name":"new","type":"sensor"}`))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandle_400_InvalidJSON(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), req("id1", `not-json`))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandle_400_ValidationError(t *testing.T) {
	setup(t, &mockSvc{updateErr: &model.ValidationError{Message: "name required"}})
	resp, _ := handle(context.Background(), req("id1", `{"name":"","type":""}`))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandle_404_NotFound(t *testing.T) {
	setup(t, &mockSvc{updateErr: &model.NotFoundError{ID: "missing"}})
	resp, _ := handle(context.Background(), req("missing", `{"name":"x","type":"y"}`))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandle_400_MissingID(t *testing.T) {
	setup(t, &mockSvc{})
	resp, _ := handle(context.Background(), events.APIGatewayProxyRequest{Body: `{"name":"x","type":"y"}`})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}