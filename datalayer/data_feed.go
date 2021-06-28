package datalayer

import (
	"fmt"
	"log"
	"time"

	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type FeedItem struct {
	PK        string    `json:"PK"`
	SK        string    `json:"SK"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Caption   string    `json:"caption"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type FeedRepository interface {
	Create(id, email, caption, url string) (models.Feed, error)
}

type dynamoDBFeedRepository struct {
	Service *dynamodb.DynamoDB
}

func NewFeedRepository(sess *session.Session) FeedRepository {
	return &dynamoDBFeedRepository{
		Service: dynamodb.New(sess),
	}
}

func (r *dynamoDBFeedRepository) Create(id, email, caption, url string) (models.Feed, error) {
	pk := fmt.Sprintf("USER#%s", email)
	sk := fmt.Sprintf("PHOTO#%s", email)
	now := time.Now()
	item := FeedItem{
		PK:        pk,
		SK:        sk,
		ID:        id,
		Caption:   caption,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
		Type:      "PHOTO",
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Printf("Got error marshalling map: %v", err)
		return models.Feed{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(TableName),
	}

	_, err = r.Service.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %v", err)
		return models.Feed{}, err
	}

	return models.Feed{
		ID:        id,
		Caption:   caption,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
