package lambdahelper

import (
	"context"
	"device-registry/pkg/model"
	"device-registry/pkg/repository"
	"device-registry/pkg/service"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type DeviceService interface {
	GetDeviceByID(ctx context.Context, deviceID string) (*model.Device, error)
	SaveDevice(ctx context.Context, device *model.Device) error
	DeleteDevice(ctx context.Context, deviceID string) error
	GetAllDevices(ctx context.Context, limit int, nextToken string) (*model.PageResult, error)
	UpdateDevice(ctx context.Context, device *model.Device) error
}

func NewService() (DeviceService, error) {
	repo, err := repository.NewDeviceRepository()
	if err != nil {
		return nil, err
	}
	return service.NewDeviceService(repo), nil
}

func RequestID(ctx context.Context) string {
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		return lc.AwsRequestID
	}
	return "local"
}

func JSONResponse(statusCode int, body any) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}
}

func ErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	return JSONResponse(statusCode, map[string]string{"error": message})
}