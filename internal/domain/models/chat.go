package models

import "time"

type Chat struct {
	ID           int64     `json:"id"`
	FirstUserID  int64     `json:"first_user_id"`
	SecondUserID int64     `json:"second_user_id"`
	LastMessage  string    `json:"last_message"`
	UpdatedAt    time.Time `json:"updated_at"`
}
