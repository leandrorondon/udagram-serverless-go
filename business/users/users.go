package users

import (
	"context"
	"errors"
	"log"

	"github.com/leandrorondon/udagram-serverless-go/business/auth"
	"github.com/leandrorondon/udagram-serverless-go/datalayer"
	"github.com/leandrorondon/udagram-serverless-go/models"
)

type Service interface {
	Create(ctx context.Context, email, password string) (models.User, string, error)
	Login(ctx context.Context, email, password string) (models.User, string, error)
}

type service struct {
	repository datalayer.UserRepository
	jwtSecret  []byte
}

func NewService(r datalayer.UserRepository, secret []byte) Service {
	return &service{
		repository: r,
		jwtSecret:  secret,
	}
}

func (s *service) Create(ctx context.Context, email, password string) (models.User, string, error) {
	var user models.User
	existing, err := s.repository.Get(ctx, email)
	if err != nil {
		log.Panicf("Error getting user")
	}
	if existing != nil {
		return user, "", errors.New("user may already exist")
	}
	user, err = s.repository.Create(ctx, email, auth.HashAndSalt(password))
	if err != nil {
		log.Panicf("Error creating user")
	}
	return user, auth.GenerateToken(email, s.jwtSecret), nil
}

func (s *service) Login(ctx context.Context, email, password string) (models.User, string, error) {
	var user models.User
	existing, err := s.repository.Get(ctx, email)
	if err != nil {
		log.Panicf("Error getting user")
	}
	if existing == nil {
		log.Printf("User %s doesn't exist", email)
		return user, "", errors.New("unauthorized")
	}
	if !auth.ComparePasswords(existing.PasswordHash, password) {
		return user, "", errors.New("unauthorized")
	}

	return *existing, auth.GenerateToken(email, s.jwtSecret), nil
}
