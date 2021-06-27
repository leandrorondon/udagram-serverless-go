package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/leandrorondon/udagram-serverless-go/business/users"
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

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type NewUserResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
	Auth bool `json:"auth"`
}

type handler struct {
	service users.Service
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func (h *handler) Handler(request events.APIGatewayProxyRequest) (Response, error) {
	var req LoginRequest

	// Unmarshal the json, return 400 if error
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Fields validation
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		return Response{
			Body:       fmt.Sprintf("validatior error:  %v", err),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Create user
	user, jwt, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	// Build and return the response
	response := NewUserResponse{
		Auth: true,
		Token: jwt,
		User:  user,
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
	r := datalayer.NewUserRepository(sess)
	svc := users.NewService(r, sess)
	h := handler{svc}
	log.Println("Initializing login lambda function")
	lambda.Start(h.Handler)
}
