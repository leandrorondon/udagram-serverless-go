package models

import "time"

type User struct {
	Email string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}