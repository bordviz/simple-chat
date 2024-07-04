package dto

import (
	"fmt"
	"simple-chat/internal/validator"
	"strings"
	"time"
)

type Message struct {
	ChatID    int64     `json:"chat_id" validate:"required"`
	Sender    int64     `json:"sender" validate:"required"`
	Text      string    `json:"text" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Message) Validate() error {
	m.Text = strings.TrimSpace(m.Text)

	if err := validator.Validate(m); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}

type MessageRequest struct {
	ChatID int64  `json:"chat_id" validate:"required"`
	Text   string `json:"text" validate:"required"`
}

func (r *MessageRequest) Validate() error {
	r.Text = strings.TrimSpace(r.Text)

	if err := validator.Validate(r); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}
