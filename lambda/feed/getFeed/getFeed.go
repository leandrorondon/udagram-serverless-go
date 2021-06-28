package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/leandrorondon/udagram-serverless-go/business/feed"
	"github.com/leandrorondon/udagram-serverless-go/datalayer"
	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type handler struct {
	service   feed.Service
	jwtSecret []byte
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func (h *handler) Handler(request events.APIGatewayProxyRequest) (Response, error) {
	items, err := h.service.ListFeed()
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	type FeedResponse struct {
		Count int `json:"count"`
		Rows []models.Feed `json:"rows"`
	}


	// Build and return the response
	r := FeedResponse {
		Count: len(items),
		Rows: items,
	}
	var buf bytes.Buffer
	body, err := json.Marshal(r)
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
	log.Println("Initializing getFeed lambda function")
	lambda.Start(h.Handler)
}
