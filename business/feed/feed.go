package feed

import (
	"log"

	"github.com/leandrorondon/udagram-serverless-go/datalayer"
	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/google/uuid"
)

type Service interface {
	Create(email, caption string) (models.Feed, string, error)
	ListFeed() ([]models.Feed, error)
}

type service struct {
	repository datalayer.FeedRepository
	file       datalayer.FileRepository
}

func NewService(r datalayer.FeedRepository, f datalayer.FileRepository) Service {
	return &service{
		repository: r,
		file:       f,
	}
}

func (s *service) Create(email, caption string) (models.Feed, string, error) {
	// Create a S3 signed URL for image upload, valid for 5 minutes
	newUUID := uuid.New().String()
	signedURL, err := s.file.GetSignedURL(newUUID)
	if err != nil {
		log.Panic("Error generating Signed URL")
	}

	// Save the feed item
	url := s.file.BuildImageURL(newUUID)
	feed, err := s.repository.Create(newUUID, email, caption, url)
	if err != nil {
		log.Panic("Error creating feed")
	}

	return feed, signedURL, nil
}

func (s *service) ListFeed() ([]models.Feed, error) {
	log.Println("ListFeed")
	items, err := s.repository.ListFeed()
	if err != nil {
		log.Panic("Error listing feed")
	}

	return items, nil
}
