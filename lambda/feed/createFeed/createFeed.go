package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/leandrorondon/udagram-serverless-go/business/auth"
	"github.com/leandrorondon/udagram-serverless-go/business/feed"
	"github.com/leandrorondon/udagram-serverless-go/datalayer"
	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-playground/validator"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type NewFeedRequest struct {
	Caption string `json:"caption" validate:"required"`
}

type NewFeedResponse struct {
	SignedURL string      `json:"signed_url"`
	Item      models.Feed `json:"item"`
}

type handler struct {
	service   feed.Service
	jwtSecret []byte
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func (h *handler) Handler(request events.APIGatewayProxyRequest) (Response, error) {
	var newItem NewFeedRequest

	// Unmarshal the json, return 400 if error
	err := json.Unmarshal([]byte(request.Body), &newItem)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Fields validation
	validate := validator.New()
	err = validate.Struct(newItem)
	if err != nil {
		return Response{
			Body:       fmt.Sprintf("validation error:  %v", err),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Get user's email from request
	email, err := auth.GetEmailFromRequest(request, h.jwtSecret)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Create feed
	item, signedURL, err := h.service.Create(email, newItem.Caption)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Build and return the response
	response := NewFeedResponse{
		Item:      item,
		SignedURL: signedURL,
	}
	var buf bytes.Buffer
	body, err := json.Marshal(response)
	if err != nil {
		log.Panicf("Error marshaling user object: %v", err)
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      http.StatusOK,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
			"Content-Type":                     "application/json",
		},
	}
	return resp, nil
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	r := datalayer.NewFeedRepository(sess)
	f := datalayer.NewFileRepository(sess)
	jwtSecret := datalayer.GetJwtSecret(sess)
	svc := feed.NewService(r, f)
	h := handler{
		service:   svc,
		jwtSecret: jwtSecret,
	}
	log.Println("Initializing createFeed lambda function")
	lambda.Start(h.Handler)
}
