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
	ListFeed() ([]models.Feed, error)
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
	sk := fmt.Sprintf("PHOTO#%s", id)
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

func (r *dynamoDBFeedRepository) ListFeed() ([]models.Feed, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String(PhotosIdx),

		KeyConditionExpression: aws.String("#itemType = :type"),
		ExpressionAttributeNames: map[string]*string{
			"#itemType": aws.String("type"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":type": {
				S: aws.String("PHOTO"),
			},
		},
		ScanIndexForward: aws.Bool(false), // sort by sort key in DESC order
	}

	result, err := r.Service.Query(input)
	if err != nil {
		log.Printf("Got error calling Query: %v", err)
		return nil, err
	}
	log.Printf("Got %d items", *result.Count)

	var feedItems []FeedItem
	if *result.Count > 0 {
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &feedItems)
		if err != nil {
			log.Printf("Got error unmarshaling items: %v", err)
			return nil, err
		}
	}

	return itemsToModel(feedItems), nil
}

func itemsToModel(feedItems []FeedItem) []models.Feed {
	var feed []models.Feed
	for _, f := range feedItems {
		feed = append(feed, models.Feed{
			ID: f.ID,
			Caption: f.Caption,
			URL: f.URL,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		})
	}
	return feed
}
