package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/leandrorondon/udagram-serverless-go/models"
	"log"
	"net/http"

	"github.com/leandrorondon/udagram-serverless-go/business/users"
	"github.com/leandrorondon/udagram-serverless-go/datalayer"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type NewUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type NewUserResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type handler struct {
	service users.Service
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func (h *handler) Handler(request events.APIGatewayProxyRequest) (Response, error) {
	var newUser NewUserRequest

	// Unmarshal the json, return 400 if error
	err := json.Unmarshal([]byte(request.Body), &newUser)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Fields validation
	validate := validator.New()
	err = validate.Struct(newUser)
	if err != nil {
		return Response{
			Body:       fmt.Sprintf("validatior error:  %v", err),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Create user
	user, jwt, err := h.service.Create(newUser.Email, newUser.Password)
	if err != nil {
		return Response{
			Body:       err.Error(),
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Build and return the response
	response := NewUserResponse{
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
	r := datalayer.NewUserRepository()
	h := handler{
		service: users.NewService(r),
	}
	lambda.Start(h.Handler)
}
