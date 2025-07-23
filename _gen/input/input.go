//go:generate go run ../main.go && go fmt ./input_validation_gen.go
package myapp

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type PasswordInput string

func (p PasswordInput) ToValidated() (string, error) {
	if p == "" {
		return "", fmt.Errorf("password cannot be empty for hashing")
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

type DateOfBirthInput string

func (d DateOfBirthInput) ToValidated() (time.Time, error) {
	const layout = "2006-01-02" // YYYY-MM-DD
	t, err := time.Parse(layout, string(d))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date '%s' with format '%s': %w", d, layout, err)
	}
	return t, nil
}

var customTransformTypeMap = map[string]struct {
	DomainFieldName string
	DomainFieldType string
	RequiredImport  string // E.g., "time" for time.Time
}{
	"myapp.PasswordInput":    {"HashedPassword", "string", "golang.org/x/crypto/bcrypt"},
	"myapp.DateOfBirthInput": {"DateOfBirth", "time.Time", "time"},
	"myapp.TagsInput":        {"Tags", "[]string", ""}, // TagsInput.ToValidated returns []string, no extra import needed beyond myapp
	// Add more mappings here for other custom input types
}

type UserInput struct {
	Email          string           `json:"email" validate:"required,email"`
	Password       PasswordInput    `json:"password" validate:"required,min=8"` // Uses custom type
	DisplayName    string           `json:"displayName" validate:"required,min=2,max=50"`
	DateOfBirthStr DateOfBirthInput `json:"dateOfBirth" validate:"required"`
	CreatedAt      string           `json:"createdAt" validate:"required" transform:"parseTime:CreatedAt"` // Parse string to time.Time
	UnusedField    string           `json:"unused" transform:"omit"`                                       // Field to omit from ValidatedUser
}

type ProductInput struct {
	Name        string  `json:"name" validate:"required,min=5"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Description string  `json:"description"`
}
