package models

import "time"

type Feed struct {
	ID string `json:"id"`
	Caption string `json:"caption"`
	URL string `json:"url"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}