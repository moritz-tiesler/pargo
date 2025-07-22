//go:generate go run ./_gen/main.go && go fmt ./input_validation_gen.go
package main

import "fmt"

type UserInput struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8" ` // Custom transform tag
	DisplayName string `json:"displayName" validate:"required,min=2,max=50"`
	CreatedAt   string `json:"createdAt" validate:"required" transform:"parseTime:CreatedAt"` // Parse string to time.Time
	UnusedField string `json:"unused" transform:"omit"`                                       // Field to omit from ValidatedUser
}

type ProductInput struct {
	Name        string  `json:"name" validate:"required,min=5"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Description string  `json:"description"`
}

func main() {
	testProduct := ProductInput{
		Name:        "nike",
		Price:       -1.0,
		Description: "bad product",
	}

	if _, err := testProduct.ToProductInputValidated(); err != nil {
		fmt.Printf("failed to validate %v, err=%s\n", testProduct, err)
	}

	testProduct = ProductInput{
		Name:        "Snickers",
		Price:       2.0,
		Description: "lecker",
	}
	if _, err := testProduct.ToProductInputValidated(); err != nil {
		fmt.Printf("failed to validate %v, err=%s\n", testProduct, err)
	}
}
