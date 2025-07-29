# pargo

A package for generating new struct definitions for types that use [**validator**](https://github.com/go-playground/validator) tags.
Validate once and use the new types that encode a successful validation.


##  Installation

```bash
go get github.com/moritz-tiesler/pargo@latest
```
 

## Example


create a go file that will call ```pargo```. Example
```bash
$ mkdir generator
$ cd generator && touch generate.go
```

```go
// generator/generate.go
package main

import (
	"fmt"
	"io"
	"log"

	"os"

	"github.com/moritz-tiesler/pargo/pkg/generator"
)

func main() {
	gen := generator.Generator{}

	fmt.Println("invoking pargo generator")
    newFile,err := gen.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("file written: %s\n", newFile)
}

```

create a file with your type definitions

```bash
$ mkdir models
$ cd models && touch user.go
```
```go
// models/user.go
package models

import (
	"time"
)

// point the go generate command to your generate.go file
//go:generate go run ../generator/generate.go
type Userstruct struct {
	Email          string    `json:"email" validate:"required,email"`
	Password       string    `json:"password" validate:"required,min=8"`
	DisplayName    string    `json:"displayName" validate:"required,min=2,max=50"`
	DateOfBirthStr string    `json:"dateOfBirth" validate:"required"`
	TagsStr        []string  `json:"tags"`
	SecretKey      string    `json:"-"`
	LastLoginTime  time.Time `json:"lastLogin"`
	Adress         Adress    `validate:"required"`
}

type Adress struct {
	Country string `validate:"required"`
	Street  string `validate:"required"`
	Zipcode string `validate:"required"`
}
```



run ```go generate``` in your project root

```bash
go generate ./...
```

A file will be created in the ```models/``` directory