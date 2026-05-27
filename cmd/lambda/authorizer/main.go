package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var apiSecret string

func init() {
	apiSecret = os.Getenv("API_SECRET")
}

func handle(_ context.Context, req events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: isAuthorized(req.Headers["authorization"]),
	}, nil
}

func isAuthorized(header string) bool {
	const prefix = "bearer "
	if apiSecret == "" || len(header) <= len(prefix) {
		return false
	}
	if !strings.EqualFold(header[:len(prefix)], prefix) {
		return false
	}
	return header[len(prefix):] == apiSecret
}

func main() {
	lambda.Start(handle)
}