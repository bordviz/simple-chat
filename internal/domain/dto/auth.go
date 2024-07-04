package dto

import (
	"fmt"
	"simple-chat/internal/validator"
	"strings"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

func (r *RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
	r.Name = strings.TrimSpace(r.Name)

	if err := validator.Validate(r); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r *LoginRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)

	if err := validator.Validate(r); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}
