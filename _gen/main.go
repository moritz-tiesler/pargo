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

	InputFields  []InputFieldData  // Fields of the Input struct
	DomainFields []DomainFieldData // Fields for the generated Domain struct
}

// InputFieldData represents a field in the Input struct, with its original tags and transformation intent.
type InputFieldData struct {
	FieldName   string
	FieldType   string // The actual type string (e.g., "myapp.PasswordInput")
	ValidateTag string
	JSONTag     string

	IsCustomTransformType bool // True if this field uses one of our custom input types
	// Target field name and type for the domain struct (inferred from custom type's ToValidated return)
	TargetFieldName string
	TargetFieldType string
}

// DomainFieldData represents a field in the generated Domain struct.
type DomainFieldData struct {
	FieldName string
	FieldType string
}

// Map of custom input types to their target domain field name and type
// This is where the generator "knows" about your custom transformation types.
var customTransformTypeMap = map[string]struct {
	DomainFieldName string
	DomainFieldType string
	RequiredImport  string // E.g., "time" for time.Time
}{
	"myapp.PasswordInput":    {"HashedPassword", "string", ""},
	"myapp.DateOfBirthInput": {"DateOfBirth", "time.Time", "time"},
	"main.TagsInput":         {"Tags", "[]string", ""}, // TagsInput.ToValidated returns []string, no extra import needed beyond myapp
	// Add more mappings here for other custom input types
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	inputFilePath := filepath.Join(wd, "input.go")
	generatedFileName := filepath.Join(wd, "input_validation_gen.go")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing file %s: %v", inputFilePath, err)
	}

	var allTemplateData []TemplateData
	packageImports := make(map[string]struct{})

	packageImports["fmt"] = struct{}{}
	packageImports["github.com/go-playground/validator/v10"] = struct{}{}

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
			domainTypeName := inputTypeName + "Validated" // e.g., UserInput -> ValidatedUser

			var inputFields []InputFieldData
			var domainFields []DomainFieldData

			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 { // Embedded fields or unexported fields without name
					continue
				}
				fieldName := field.Names[0].Name
				fieldType := exprToString(field.Type) // e.g., "string", "myapp.PasswordInput"
				validateTag := getTagValue(field.Tag, "validate")
				jsonTag := getTagValue(field.Tag, "json")

				isCustom := false
				targetFieldName := fieldName
				targetFieldType := fieldType

				// Check if the field's type is one of our known custom transformation types
				if transformInfo, found := customTransformTypeMap[node.Name.Name+"."+fieldType]; found {
					isCustom = true
					targetFieldName = transformInfo.DomainFieldName
					targetFieldType = transformInfo.DomainFieldType
					if transformInfo.RequiredImport != "" {
						packageImports[transformInfo.RequiredImport] = struct{}{}
					}
					// Add the package where custom input types are defined
					packageImports[node.Name.Name] = struct{}{} // Assuming custom types are in the same package as Input structs
				} else {
					// Collect imports for standard types if not already covered by custom types
					if strings.Contains(fieldType, ".") {
						parts := strings.Split(fieldType, ".")
						if len(parts) > 1 {
							if parts[0] == "time" {
								packageImports["time"] = struct{}{}
							}
							// Add other common packages here as needed
						}
					}
				}

				currentInputField := InputFieldData{
					FieldName:             fieldName,
					FieldType:             fieldType,
					ValidateTag:           validateTag,
					JSONTag:               jsonTag,
					IsCustomTransformType: isCustom,
					TargetFieldName:       targetFieldName, // Will be used by template for direct copy or custom call
					TargetFieldType:       targetFieldType, // Will be used by template for domain struct def
				}
				inputFields = append(inputFields, currentInputField)

				// Determine domain fields for the generated struct definition
				if jsonTag == "-" {
					continue // Omit field from domain struct if json:"-"
				}

				// If it's a custom transform type, use its mapped domain field name/type
				if isCustom {
					domainFields = append(domainFields, DomainFieldData{
						FieldName: currentInputField.TargetFieldName,
						FieldType: currentInputField.TargetFieldType,
					})
				} else {
					// Otherwise, direct copy for the domain struct definition
					domainFields = append(domainFields, DomainFieldData{
						FieldName: fieldName,
						FieldType: fieldType,
					})
				}
			}

			allTemplateData = append(allTemplateData, TemplateData{
				InputTypeName:  inputTypeName,
				DomainTypeName: domainTypeName,
				PackageName:    node.Name.Name,
				Imports:        packageImports,
				InputFields:    inputFields,
				DomainFields:   domainFields,
			})
		}
	}

	fmt.Println(wd)
	tmpl, err := template.ParseFiles(filepath.Join(filepath.Dir(wd), "templates", "generator_template.tmpl"))
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", node.Name.Name))

	if len(packageImports) > 0 {
		buf.WriteString("import (\n")
		sortedImports := make([]string, 0, len(packageImports))
		for pkg := range packageImports {
			sortedImports = append(sortedImports, pkg)
		}
		stdImports := []string{}
		otherImports := []string{}
		for _, imp := range sortedImports {
			if !strings.Contains(imp, ".") {
				stdImports = append(stdImports, imp)
			} else {
				otherImports = append(otherImports, imp)
			}
		}
		// strings.Sort(stdImports)
		// strings.Sort(otherImports)

		for _, imp := range stdImports {
			buf.WriteString(fmt.Sprintf("    \"%s\"\n", imp))
		}
		if len(stdImports) > 0 && len(otherImports) > 0 {
			buf.WriteString("\n")
		}
		for _, imp := range otherImports {
			buf.WriteString(fmt.Sprintf("    \"%s\"\n", imp))
		}
		buf.WriteString(")\n\n")
	}

	buf.WriteString("var validate = validator.New()\n\n")

	for _, d := range allTemplateData {
		err = tmpl.Execute(&buf, d)
		if err != nil {
			log.Fatalf("Error executing template for %s: %v", d.InputTypeName, err)
		}
	}

	formattedSource, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Failed to format generated code. Raw content:\n%s\n", buf.String())
		log.Fatalf("Error formatting generated Go code: %v", err)
	}

	err = os.WriteFile(generatedFileName, formattedSource, 0644)
	if err != nil {
		log.Fatalf("Error writing generated file: %v", err)
	}

	log.Printf("Successfully generated and formatted %s\n", generatedFileName)
}

// Helper functions (getTagValue, exprToString) remain the same.
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
