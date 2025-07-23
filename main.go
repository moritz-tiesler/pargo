package main

import (
	"fmt"
	myapp "pargo/input"
)

func main() {
	testProduct := myapp.ProductInput{
		Name:        "nike",
		Price:       -1.0,
		Description: "bad product",
	}

	if _, err := testProduct.ToProductInputValidated(); err != nil {
		fmt.Printf("failed to validate %v, err=%s\n", testProduct, err)
	}

	testProduct = myapp.ProductInput{
		Name:        "Snickers",
		Price:       2.0,
		Description: "lecker",
	}
	if _, err := testProduct.ToProductInputValidated(); err != nil {
		fmt.Printf("failed to validate %v, err=%s\n", testProduct, err)
	}
}
