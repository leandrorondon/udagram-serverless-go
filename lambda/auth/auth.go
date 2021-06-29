package main

import (
	"errors"
	"github.com/leandrorondon/udagram-serverless-go/business/auth"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/leandrorondon/udagram-serverless-go/datalayer"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type handler struct {
	jwtSecret []byte
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func (h *handler) Handler(request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := auth.ParseTokenFromRequest(request, h.jwtSecret)
	if token == nil || !token.Valid {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
	email, err := auth.GetEmailFromToken(token)
	if err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	return auth.GeneratePolicy(email, "Allow", request.MethodArn), nil
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	h := handler{
		jwtSecret: datalayer.GetJwtSecret(sess),
	}
	log.Println("Initializing auth lambda function")
	lambda.Start(h.Handler)
}
