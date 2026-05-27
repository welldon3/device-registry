package main

import (
	"context"
	"device-registry/internal/lambdahelper"
	"device-registry/pkg/service"
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

	if err := svc.DeleteDevice(ctx, deviceID); err != nil {
		if service.IsNotFoundError(err) {
			return lambdahelper.ErrorResponse(http.StatusNotFound, err.Error()), nil
		}
		log.Printf("[%s] Error deleting device %s: %v", requestID, deviceID, err)
		return lambdahelper.ErrorResponse(http.StatusInternalServerError, "error deleting device"), nil
	}

	return lambdahelper.JSONResponse(http.StatusOK, map[string]string{"message": "device deleted successfully"}), nil
}

func main() {
	lambda.Start(handle)
}