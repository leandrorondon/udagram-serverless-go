package auth

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Panicf("Error generating hash from password: %v", err)
	}
	return string(hash)
}

func ComparePasswords(hashed string, plain string) bool {
	byteHash := []byte(hashed)
	bytePlain := []byte(plain)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePlain)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func GenerateToken(email string, secret []byte) string {
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["email"] = email
	atClaims["exp"] = time.Now().Add(time.Hour * 168).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(secret)
	if err != nil {
		log.Panicf("Error signing token: %v", err)
	}
	return token
}

func GetEmailFromRequest(request events.APIGatewayProxyRequest, secret []byte) (string, error) {
	authorization, ok := request.Headers["Authorization"]
	if !ok {
		return "", errors.New("authorization not found")
	}

	tokenString := strings.Split(authorization, " ")[1]

	type CustomClaims struct {
		Email string `json:"email"`
		jwt.StandardClaims
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.Email, nil

}
