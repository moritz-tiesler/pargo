# pargo

A package for generating new struct definitions for types that use [**validator**](https://github.com/go-playground/validator) tags.

Validate once and use the new types that encode a successful validation.


##  Installation

```bash
go get github.com/moritz-tiesler/pargo@latest
```
## Example

### Create a go file that will call ```pargo```
```bash
$ mkdir generator
$ cd generator && touch generate.go
```

```go
// generator/generate.go
package main

import (
	"fmt"
	"log"

	"github.com/moritz-tiesler/pargo/pkg/generator"
)

func main() {
	gen := generator.Generator{}

	fmt.Println("invoking pargo generator")
	newFile, err := gen.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("file written: %s\n", newFile)
}

```

### Create a file with your type definitions

```bash
$ mkdir models
$ cd models && touch user.go
```
```go
// models/user.go
package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// set up you validator once for the package. Its name must be 'VALIDATE'
// The generated code will reference this validator.
var VALIDATE = validator.New(validator.WithRequiredStructEnabled())

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



### Run ```go generate``` in your project root

```bash
go generate ./...
```

A file will be created in the ```models/``` directory

```bash
$ ls models
user.go  user_gen.go
```

### Use the new types definitions

```go
// main.go
package main

import (
	"fmt"
	"pargo_use/models"
)

func main() {
	u := models.Userstruct{
		DisplayName:    "Bob",
		Email:          "bob@bobnet.com",
		Password:       "s3cr3T",
		DateOfBirthStr: "01.01.2000",
		Adress: models.Adress{
			Country: "England",
		},
	}

	if uv, err := u.ToUserstructValidated(); err != nil {
		fmt.Printf("error validating %+v, err=%s\n", u, err)
	} else {
		fmt.Printf("validated user %+v", uv)
	}
}
```
