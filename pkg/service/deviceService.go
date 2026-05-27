package service

import (
	"context"
	"device-registry/pkg/model"
	"errors"
)

type DeviceRepository interface {
	GetDeviceByID(ctx context.Context, deviceID string) (*model.Device, error)
	SaveDevice(ctx context.Context, device *model.Device) error
	DeleteDevice(ctx context.Context, deviceID string) error
	GetAllDevices(ctx context.Context, limit int, nextToken string) (*model.PageResult, error)
	UpdateDevice(ctx context.Context, device *model.Device) error
}

type DeviceService struct {
	repo DeviceRepository
}

func NewDeviceService(repo DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

func (s *DeviceService) GetDeviceByID(ctx context.Context, deviceID string) (*model.Device, error) {
	return s.repo.GetDeviceByID(ctx, deviceID)
}

func (s *DeviceService) SaveDevice(ctx context.Context, device *model.Device) error {
	if device.DeviceID == "" || device.Name == "" || device.Type == "" {
		return &model.ValidationError{Message: "deviceId, name and type are required"}
	}
	if !isValidDeviceID(device.DeviceID) {
		return &model.ValidationError{Message: "deviceId must be 1-128 chars: letters, digits, hyphens and underscores only"}
	}
	if len(device.Name) > 256 {
		return &model.ValidationError{Message: "name must be 256 chars or fewer"}
	}
	if len(device.Type) > 128 {
		return &model.ValidationError{Message: "type must be 128 chars or fewer"}
	}
	return s.repo.SaveDevice(ctx, device)
}

func (s *DeviceService) DeleteDevice(ctx context.Context, deviceID string) error {
	return s.repo.DeleteDevice(ctx, deviceID)
}

func (s *DeviceService) GetAllDevices(ctx context.Context, limit int, nextToken string) (*model.PageResult, error) {
	return s.repo.GetAllDevices(ctx, limit, nextToken)
}

func (s *DeviceService) UpdateDevice(ctx context.Context, device *model.Device) error {
	if device.Name == "" || device.Type == "" {
		return &model.ValidationError{Message: "name and type are required"}
	}
	if len(device.Name) > 256 {
		return &model.ValidationError{Message: "name must be 256 chars or fewer"}
	}
	if len(device.Type) > 128 {
		return &model.ValidationError{Message: "type must be 128 chars or fewer"}
	}
	return s.repo.UpdateDevice(ctx, device)
}

func IsValidationError(err error) bool {
	var ve *model.ValidationError
	return errors.As(err, &ve)
}

func IsNotFoundError(err error) bool {
	var nfe *model.NotFoundError
	return errors.As(err, &nfe)
}

func IsConflictError(err error) bool {
	var ce *model.ConflictError
	return errors.As(err, &ce)
}

func isValidDeviceID(id string) bool {
	if len(id) > 128 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if !isASCIIAlnum(c) && c != '-' && c != '_' {
			return false
		}
	}
	return true
}

func isASCIIAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}