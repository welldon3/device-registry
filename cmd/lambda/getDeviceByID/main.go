package main

import (
	"context"
	"device-registry/internal/lambdahelper"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	telemetry "github.com/welldon3/go-telemetry"
)

var svc lambdahelper.DeviceService
var tel *telemetry.Provider

func init() {
	var err error
	svc, err = lambdahelper.NewService()
	if err != nil {
		log.Fatalf("Unable to initialize repository: %v", err)
	}

	tel, err = telemetry.New(context.Background(), telemetry.Config{
		ServiceName: "device-registry",
		Exporter:    telemetry.ExporterStdout,
	})
	if err != nil {
		log.Fatalf("Unable to initialize telemetry: %v", err)
	}
}

func handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctx, span := tel.Tracer().Start(ctx, "getDeviceByID")
	defer span.End()

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