package datalayer

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var (
	JwtSecretID    = os.Getenv("JWT_SECRET_ID")
	JwtSecretField = os.Getenv("JWT_SECRET_FIELD")
)

func GetJwtSecret(sess *session.Session) []byte {
	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(JwtSecretID),
	}
	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.Panicf("error getting secret: %v", err)
	}
	if result.SecretString == nil {
		log.Panic("nil SecretString")
	}

	m := make(map[string]string)
	err = json.Unmarshal([]byte(*result.SecretString), &m)
	if err != nil {
		log.Panicf("Error unmarshaling secret: %v", err)
	}

	return []byte(m[JwtSecretField])
}