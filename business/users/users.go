package users

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/leandrorondon/udagram-serverless-go/datalayer"
	"github.com/leandrorondon/udagram-serverless-go/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(string, string) (models.User, string, error)
	Login(string, string) (models.User, string, error)
}

type service struct {
	repository  datalayer.UserRepository
	jwtSecret []byte
}
var (
	JwtSecretID = os.Getenv("JWT_SECRET_ID")
	JwtSecretField = os.Getenv("JWT_SECRET_FIELD")
)

func NewService(r datalayer.UserRepository, sess *session.Session) Service {
	jwtSecret := getJwtSecret(sess)

	return &service{
		repository:  r,
		jwtSecret: jwtSecret,
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
	user, err = s.repository.Create(email, s.hashAndSalt(password))
	if err != nil {
		log.Panicf("Error creating user")
	}
	return user, s.generateToken(email), nil
}

func (s *service) Login(email, password string) (models.User, string, error) {
	var user models.User
	existing, err := s.repository.Get(email)
	if err != nil {
		log.Panicf("Error getting user")
	}
	if existing == nil {
		log.Printf("User %s doesn't exist", email)
		return user, "", errors.New("unauthorized")
	}
	if !s.comparePasswords(existing.PasswordHash, password) {
		return user, "", errors.New("unauthorized")
	}

	return *existing, s.generateToken(email), nil
}

func (s *service) hashAndSalt(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Panicf("Error generating hash from password: %v", err)
	}
	return string(hash)
}

func (s *service) comparePasswords(hashed string, plain string) bool {
	byteHash := []byte(hashed)
	bytePlain := []byte(plain)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePlain)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (s *service) generateToken(email string) string {
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["email"] = email
	atClaims["exp"] = time.Now().Add(time.Hour * 168).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(s.jwtSecret)
	if err != nil {
		log.Panicf("Error signing token: %v", err)
	}
	return token
}

func getJwtSecret(sess *session.Session) []byte {
	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(JwtSecretID),
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