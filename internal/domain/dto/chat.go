package dto

import (
	"fmt"
	"simple-chat/internal/validator"
	"time"
)

type Chat struct {
	FirstUserID  int64     `json:"first_user_id" validate:"required"`
	SecondUserID int64     `json:"second_user_id" validate:"required"`
	LastMessage  string    `json:"last_message"`
	UpdatedAt    time.Time `json:"updated_at" validate:"required"`
}

func (c *Chat) Validate() error {
	if err := validator.Validate(c); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}

type CreateChatRequest struct {
	FirstUserID  int64 `json:"first_user_id" validate:"required"`
	SecondUserID int64 `json:"second_user_id" validate:"required"`
}

func (c *CreateChatRequest) Validate() error {
	if err := validator.Validate(c); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}
