package datalayer

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	TableName = os.Getenv("UDAGRAM_TABLE")
)

type UserItem struct {
	PK           string    `json:"PK"`
	SK           string    `json:"SK"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type UserRepository interface {
	Create(string, string) (models.User, error)
	Get(string) (*models.User, error)
}

type dynamoDBUserRepository struct {
	Service *dynamodb.DynamoDB
}

func NewUserRepository(sess *session.Session) UserRepository {
	return &dynamoDBUserRepository{
		Service: dynamodb.New(sess),
	}
}

func (r *dynamoDBUserRepository) Create(email string, passwordHash string) (models.User, error) {
	pk := fmt.Sprintf("USER#%s", email)
	sk := fmt.Sprintf("METADATA#%s", email)
	now := time.Now()
	item := UserItem{
		PK:           pk,
		SK:           sk,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Printf("Got error marshalling map: %v", err)
		return models.User{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(TableName),
	}

	_, err = r.Service.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %v", err)
		return models.User{}, err
	}

	return models.User{
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (r *dynamoDBUserRepository) Get(email string) (*models.User, error) {
	pk := fmt.Sprintf("USER#%s", email)
	sk := fmt.Sprintf("METADATA#%s", email)

	result, err := r.Service.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(pk),
			},
			"SK": {
				S: aws.String(sk),
			},
		},
	})

	if err != nil {
		log.Printf("Got error calling GetItem: %v", err)
		return nil, err
	}
	if result.Item == nil {
		log.Printf("Could not find user with email %s", email)
		return nil, nil
	}

	var item UserItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		log.Printf("Failed to unmarshal Record, %v", err)
		return nil, err
	}

	user := models.User{
		Email:        item.Email,
		PasswordHash: item.PasswordHash,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}

	return &user, nil
}
