package main

import (
	"context"
	"device-registry/internal/lambdahelper"
	"device-registry/pkg/model"
	"device-registry/pkg/service"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var svc lambdahelper.DeviceService

func init() {
	var err error
	svc, err = lambdahelper.NewService()
	if err != nil {
		log.Fatalf("Unable to initialize repository: %v", err)
	}
}

func handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestID := lambdahelper.RequestID(ctx)

	deviceID, ok := request.PathParameters["id"]
	if !ok || deviceID == "" {
		return lambdahelper.ErrorResponse(http.StatusBadRequest, "missing device ID"), nil
	}

	var device model.Device
	if err := json.Unmarshal([]byte(request.Body), &device); err != nil {
		return lambdahelper.ErrorResponse(http.StatusBadRequest, "invalid request body"), nil
	}
	device.DeviceID = deviceID

	if err := svc.UpdateDevice(ctx, &device); err != nil {
		if service.IsValidationError(err) {
			return lambdahelper.ErrorResponse(http.StatusBadRequest, err.Error()), nil
		}
		if service.IsNotFoundError(err) {
			return lambdahelper.ErrorResponse(http.StatusNotFound, err.Error()), nil
		}
		log.Printf("[%s] Error updating device %s: %v", requestID, deviceID, err)
		return lambdahelper.ErrorResponse(http.StatusInternalServerError, "error updating device"), nil
	}

	return lambdahelper.JSONResponse(http.StatusOK, &device), nil
}

func main() {
	lambda.Start(handle)
}