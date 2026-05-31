package main

import (
	"context"
	"device-registry/internal/lambdahelper"
	"device-registry/pkg/service"
	"log"
	"net/http"
	"strconv"

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
	ctx, span := tel.Tracer().Start(ctx, "getAllDevices")
	defer span.End()

	requestID := lambdahelper.RequestID(ctx)

	limit := 100
	if raw := request.QueryStringParameters["limit"]; raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 1000 {
			return lambdahelper.ErrorResponse(http.StatusBadRequest, "limit must be an integer between 1 and 1000"), nil
		}
		limit = parsed
	}
	nextToken := request.QueryStringParameters["nextToken"]

	result, err := svc.GetAllDevices(ctx, limit, nextToken)
	if err != nil {
		if service.IsValidationError(err) {
			return lambdahelper.ErrorResponse(http.StatusBadRequest, err.Error()), nil
		}
		log.Printf("[%s] Error getting all devices: %v", requestID, err)
		return lambdahelper.ErrorResponse(http.StatusInternalServerError, "error retrieving devices"), nil
	}

	return lambdahelper.JSONResponse(http.StatusOK, result), nil
}

func main() {
	lambda.Start(handle)
}