//go:generate go run ../main.go && go fmt ./input_validation_gen.go

package myapp

import (
	"time"
)

type UserInput struct {
	Email          string    `json:"email" validate:"required,email"`
	Password       string    `json:"password" validate:"required,min=8"` // Back to string
	DisplayName    string    `json:"displayName" validate:"required,min=2,max=50"`
	DateOfBirthStr string    `json:"dateOfBirth" validate:"required"` // Back to string
	TagsStr        []string  `json:"tags"`                            // Back to []string
	SecretKey      string    `json:"-"`                               // Still omitted by convention
	LastLoginTime  time.Time `json:"lastLogin"`                       // Direct copy
}

type ProductInput struct {
	Name        string  `json:"name" validate:"required,min=5"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Description string  `json:"description"`
}
