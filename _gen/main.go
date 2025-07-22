package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData holds the information needed to generate code for one Input struct.
type TemplateData struct {
	InputTypeName  string
	DomainTypeName string
	PackageName    string
	Imports        map[string]struct{} // Collect unique imports needed by generated code

	InputFields  []FieldData // Fields of the Input struct
	DomainFields []FieldData // Fields for the generated Domain struct
}

// FieldData represents a field in either the Input or Domain struct.
type FieldData struct {
	FieldName string
	FieldType string
	// Tags relevant to the Input struct for parsing/validation
	ValidateTag string
	JSONTag     string

	// Specific flags for transformations derived by convention
	NeedsHashing bool
}

func main() {
	// Determine the current working directory to find the input file
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	log.Printf("cwd=%s\n", wd)

	// Assuming the input file is in the parent directory (e.g., myapp/input.go)
	// You might need to adjust this path based on your project structure.
	inputFilePath := filepath.Join(wd, "input.go")
	generatedFileName := filepath.Join(wd, "input_validation_gen.go")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing file %s: %v", inputFilePath, err)
	}

	var allTemplateData []TemplateData
	packageImports := make(map[string]struct{}) // To collect all imports across all structs

	// AST traversal to find struct types with "Input" suffix
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok || !strings.HasSuffix(typeSpec.Name.Name, "Input") {
				continue
			}

			inputTypeName := typeSpec.Name.Name
			domainTypeName := inputTypeName + "Validated" // e.g., UserInput -> UserInputValidated

			// Prepare field data for both Input and Domain structs
			var inputFields []FieldData
			var domainFields []FieldData

			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 { // Embedded fields or unexported fields without name
					continue
				}
				fieldName := field.Names[0].Name
				fieldType := exprToString(field.Type) // Helper to convert ast.Expr to string
				validateTag := getTagValue(field.Tag, "validate")
				jsonTag := getTagValue(field.Tag, "json")

				currentInputField := FieldData{
					FieldName:   fieldName,
					FieldType:   fieldType,
					ValidateTag: validateTag,
					JSONTag:     jsonTag,
				}
				inputFields = append(inputFields, currentInputField)

				// Apply conventions to determine domain fields
				if jsonTag == "-" { // Convention 1: Omit if json:"-"
					continue
				}

				if fieldName == "Password" && strings.Contains(validateTag, "min=") {
					// Convention 2: If field is "Password" and has a min length validation,
					// assume it needs hashing and rename for domain struct.
					domainFields = append(domainFields, FieldData{
						FieldName:    "Password",
						FieldType:    "string", // Hashed passwords are strings
						NeedsHashing: true,
					})
					// Add import for bcrypt if you were actually hashing
					// packageImports["golang.org/x/crypto/bcrypt"] = struct{}{}
				} else if fieldType == "string" && strings.Contains(validateTag, "email") {
					// Convention 3: Specific types/validations map directly (demonstrative)
					domainFields = append(domainFields, FieldData{
						FieldName: fieldName,
						FieldType: fieldType,
					})
				} else {
					// Default: Direct copy
					domainFields = append(domainFields, FieldData{
						FieldName: fieldName,
						FieldType: fieldType,
					})
				}

				// Collect imports based on field types (basic example)
				if strings.Contains(fieldType, ".") { // e.g., "time.Time"
					parts := strings.Split(fieldType, ".")
					if len(parts) > 1 {
						// This is a very simplistic way to guess imports.
						// A more robust solution would use go/types for accurate package paths.
						if parts[0] == "time" {
							packageImports["time"] = struct{}{}
						}
						// Add other common packages here as needed
					}
				}
			}

			allTemplateData = append(allTemplateData, TemplateData{
				InputTypeName:  inputTypeName,
				DomainTypeName: domainTypeName,
				PackageName:    node.Name.Name, // Get package name from parsed file
				Imports:        packageImports,
				InputFields:    inputFields,
				DomainFields:   domainFields,
			})
		}
	}

	// Load and execute template
	tmpl, err := template.ParseFiles(filepath.Join(wd, "_gen", "templates", "generator_template.tmpl"))
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", node.Name.Name)) // Use the actual package name

	// Write imports
	if len(packageImports) > 0 {
		buf.WriteString("import (\n")
		buf.WriteString("    \"fmt\"\n") // Always needed for fmt.Errorf
		buf.WriteString("    \"github.com/go-playground/validator/v10\"\n")
		// Add other common imports here if always needed, or iterate packageImports
		for pkg := range packageImports {
			buf.WriteString(fmt.Sprintf("    \"%s\"\n", pkg))
		}
		buf.WriteString(")\n\n")
	} else {
		buf.WriteString("import (\n")
		buf.WriteString("    \"fmt\"\n\n")
		buf.WriteString("    \"github.com/go-playground/validator/v10\"\n")
		buf.WriteString(")\n\n")
	}

	buf.WriteString("var validate = validator.New()\n")

	for _, d := range allTemplateData {
		err = tmpl.Execute(&buf, d)
		if err != nil {
			log.Fatalf("Error executing template for %s: %v", d.InputTypeName, err)
		}
	}

	// Format code
	source, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Failed to format generated code. Raw content:\n%s\n", buf.String())
		log.Fatalf("Error formatting generated Go code: %v", err)
	}
	// Write to file
	err = os.WriteFile(generatedFileName, source, 0644)
	if err != nil {
		log.Fatalf("Error writing generated file: %v", err)
	}

	log.Printf("Successfully generated %s\n", generatedFileName)
}

// Helper to extract tag value
func getTagValue(tag *ast.BasicLit, key string) string {
	if tag == nil {
		return ""
	}
	s := strings.Trim(tag.Value, "`")
	parts := strings.Fields(s)
	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			val := strings.TrimPrefix(part, key+":")
			return strings.Trim(val, "\"")
		}
	}
	return ""
}

// exprToString converts an ast.Expr (type expression) to its string representation.
// This is a basic implementation and might need to be expanded for complex types
// (e.g., nested structs, interfaces, chan types).
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	default:
		return fmt.Sprintf("interface{} /* UnsupportedType: %T */", expr)
	}
}
