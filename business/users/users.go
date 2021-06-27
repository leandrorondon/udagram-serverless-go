package users

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/leandrorondon/udagram-serverless-go/datalayer"
	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(string, string) (models.User, string, error)
}

type service struct {
	repository datalayer.UserRepository
}

func NewService(r datalayer.UserRepository) Service {
	return &service{
		repository: r,
	}
}

func (s *service) Create(email, password string) (models.User, string, error) {
	var user models.User
	existing, err := s.repository.Get(email)
	if err != nil {
		log.Panicf("Error getting user")
	}
	if existing != nil {
		return user, "", errors.New("user may already exist")
	}
	user, err = s.repository.Create(email, hashAndSalt(password))
	if err != nil {
		log.Panicf("Error creating user")
	}
	return user, generateToken(email), nil
}

func hashAndSalt(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Panicf("Error generating hash from password: %v", err)
	}
	return string(hash)
}

func comparePasswords(hashed string, plain string) bool {
	byteHash := []byte(hashed)
	bytePlain := []byte(plain)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePlain)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func generateToken(email string) string {
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["email"] = email
	atClaims["exp"] = time.Now().Add(time.Hour * 168).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Panicf("Error signing token: %v", err)
	}
	return token
}
