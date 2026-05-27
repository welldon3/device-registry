package main

import (
	"context"
	"device-registry/internal/lambdahelper"
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

	device, err := svc.GetDeviceByID(ctx, deviceID)
	if err != nil {
		log.Printf("[%s] Error getting device %s: %v", requestID, deviceID, err)
		return lambdahelper.ErrorResponse(http.StatusInternalServerError, "error retrieving device"), nil
	}
	if device == nil {
		return lambdahelper.ErrorResponse(http.StatusNotFound, "device not found"), nil
	}

	return lambdahelper.JSONResponse(http.StatusOK, device), nil
}

func main() {
	lambda.Start(handle)
}