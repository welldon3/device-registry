package service_test

import (
	"context"
	"device-registry/pkg/model"
	"device-registry/pkg/service"
	"testing"
)

type mockRepo struct {
	devices map[string]*model.Device
}

func newMockRepo() *mockRepo {
	return &mockRepo{devices: make(map[string]*model.Device)}
}

func (m *mockRepo) GetDeviceByID(_ context.Context, id string) (*model.Device, error) {
	d, ok := m.devices[id]
	if !ok {
		return nil, nil
	}
	return d, nil
}

func (m *mockRepo) SaveDevice(_ context.Context, d *model.Device) error {
	if _, exists := m.devices[d.DeviceID]; exists {
		return &model.ConflictError{ID: d.DeviceID}
	}
	m.devices[d.DeviceID] = d
	return nil
}

func (m *mockRepo) DeleteDevice(_ context.Context, id string) error {
	if _, exists := m.devices[id]; !exists {
		return &model.NotFoundError{ID: id}
	}
	delete(m.devices, id)
	return nil
}

func (m *mockRepo) GetAllDevices(_ context.Context, _ int, _ string) (*model.PageResult, error) {
	devices := make([]model.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, *d)
	}
	return &model.PageResult{Devices: devices}, nil
}

func (m *mockRepo) UpdateDevice(_ context.Context, d *model.Device) error {
	if _, exists := m.devices[d.DeviceID]; !exists {
		return &model.NotFoundError{ID: d.DeviceID}
	}
	m.devices[d.DeviceID] = d
	return nil
}

var ctx = context.Background()

func TestSaveDevice_ValidationError(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	if err := svc.SaveDevice(ctx, &model.Device{}); !service.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestSaveDevice_MissingFields(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	cases := []model.Device{
		{Name: "x", Type: "y"},
		{DeviceID: "id1", Type: "y"},
		{DeviceID: "id1", Name: "x"},
	}
	for _, d := range cases {
		d := d
		if err := svc.SaveDevice(ctx, &d); !service.IsValidationError(err) {
			t.Fatalf("expected ValidationError for %+v, got %v", d, err)
		}
	}
}

func TestSaveDevice_InvalidDeviceID(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	cases := []string{"has space", "has@at", "has/slash", "x" + string(make([]byte, 129))}
	for _, id := range cases {
		d := &model.Device{DeviceID: id, Name: "n", Type: "t"}
		if err := svc.SaveDevice(ctx, d); !service.IsValidationError(err) {
			t.Fatalf("expected ValidationError for deviceId %q, got %v", id, err)
		}
	}
}

func TestSaveDevice_Conflict(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	d := &model.Device{DeviceID: "id1", Name: "router", Type: "network"}
	_ = svc.SaveDevice(ctx, d)
	if err := svc.SaveDevice(ctx, d); !service.IsConflictError(err) {
		t.Fatalf("expected ConflictError, got %v", err)
	}
}

func TestSaveDevice_OK(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	if err := svc.SaveDevice(ctx, &model.Device{DeviceID: "id1", Name: "router", Type: "network"}); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDevice_ValidationError(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	if err := svc.UpdateDevice(ctx, &model.Device{DeviceID: "id1"}); !service.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestUpdateDevice_NotFound(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	if err := svc.UpdateDevice(ctx, &model.Device{DeviceID: "missing", Name: "x", Type: "y"}); !service.IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestUpdateDevice_OK(t *testing.T) {
	repo := newMockRepo()
	repo.devices["id1"] = &model.Device{DeviceID: "id1", Name: "old", Type: "sensor"}
	svc := service.NewDeviceService(repo)
	if err := svc.UpdateDevice(ctx, &model.Device{DeviceID: "id1", Name: "new", Type: "sensor"}); err != nil {
		t.Fatal(err)
	}
	if repo.devices["id1"].Name != "new" {
		t.Fatal("device was not updated")
	}
}

func TestGetDeviceByID_Found(t *testing.T) {
	repo := newMockRepo()
	repo.devices["id1"] = &model.Device{DeviceID: "id1", Name: "router", Type: "network"}
	svc := service.NewDeviceService(repo)
	d, err := svc.GetDeviceByID(ctx, "id1")
	if err != nil || d == nil || d.Name != "router" {
		t.Fatalf("unexpected result: %v %v", d, err)
	}
}

func TestGetDeviceByID_NotFound(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	d, err := svc.GetDeviceByID(ctx, "missing")
	if err != nil || d != nil {
		t.Fatalf("expected nil device, got %v %v", d, err)
	}
}

func TestDeleteDevice_NotFound(t *testing.T) {
	svc := service.NewDeviceService(newMockRepo())
	if err := svc.DeleteDevice(ctx, "missing"); !service.IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestDeleteDevice_OK(t *testing.T) {
	repo := newMockRepo()
	repo.devices["id1"] = &model.Device{DeviceID: "id1", Name: "router", Type: "network"}
	svc := service.NewDeviceService(repo)
	if err := svc.DeleteDevice(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
	if _, exists := repo.devices["id1"]; exists {
		t.Fatal("device was not deleted")
	}
}

func TestGetAllDevices(t *testing.T) {
	repo := newMockRepo()
	repo.devices["id1"] = &model.Device{DeviceID: "id1", Name: "a", Type: "x"}
	repo.devices["id2"] = &model.Device{DeviceID: "id2", Name: "b", Type: "y"}
	svc := service.NewDeviceService(repo)
	result, err := svc.GetAllDevices(ctx, 100, "")
	if err != nil || len(result.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d %v", len(result.Devices), err)
	}
}