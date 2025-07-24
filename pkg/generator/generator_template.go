package generator

const GeneratorTemplate = `
{{/*
	Template for generating ValidatedX struct definitions and ToValidatedX methods.
	Data is of type TemplateData
*/}}

// {{.DomainTypeName}} represents a validated {{.InputTypeName}}.
// Its existence guarantees that all data within it has passed initial validation rules.
type {{.DomainTypeName}} struct {
{{- range .DomainFields}}
	{{- if .NewName | ne ""}}
	{{.NewName}} {{.FieldType}} {{.Tag}}
	{{- else}}
	{{.FieldName}} {{.FieldType}} {{.Tag}}
	{{- end}}
{{- end}}
}

// To{{.DomainTypeName}} takes a {{.InputTypeName}}, validates it, and if successful,
// converts it into a {{.DomainTypeName}}.
func (input {{.InputTypeName}}) To{{.DomainTypeName}}() (*{{.DomainTypeName}}, error) {
	if err := validate.Struct(input); err != nil {
		return nil, fmt.Errorf("validation failed for {{.InputTypeName}}: %w", err)
	}

	validated := &{{.DomainTypeName}}{} // Use the generated domain type

{{- range .DomainFields}}
	{{- if .NewName | ne ""}}
	validated.{{.NewName}}= input.{{.FieldName}} // Direct copy
	{{- else}}
	validated.{{.FieldName}} = input.{{.FieldName}} // Direct copy
	{{- end}}
{{- end}}

	return validated, nil
}
`

const GeneratorCastTemplate = `
{{/*
	Template for generating ValidatedX struct definitions and ToValidatedX methods.
	Data is of type TemplateData
*/}}

// {{.DomainTypeName}} represents a validated {{.InputTypeName}}.
// Its existence guarantees that all data within it has passed initial validation rules.
type {{.DomainTypeName}} struct {
{{- range .DomainFields}}
	{{- if .NewName | ne ""}}
	{{.NewName}} {{.NewName}}Valid{{.Tag}}
	{{- else}}
	{{.FieldName}} {{.FieldName}}Valid{{.Tag}}
	{{- end}}
{{- end}}
}

// To{{.DomainTypeName}} takes a {{.InputTypeName}}, validates it, and if successful,
// converts it into a {{.DomainTypeName}}.
func (input {{.InputTypeName}}) To{{.DomainTypeName}}() (*{{.DomainTypeName}}, error) {
	if err := validate.Struct(input); err != nil {
		return nil, fmt.Errorf("validation failed for {{.InputTypeName}}: %w", err)
	}

	validated := &{{.DomainTypeName}}{} // Use the generated domain type

{{- range .DomainFields}}
	{{- if .NewName | ne ""}}
	validated.{{.NewName}}= {{.NewName}}Valid(input.{{.FieldName}}) // Direct copy
	{{- else}}
	validated.{{.FieldName}} = {{.FieldName}}Valid(input.{{.FieldName}}) // Direct copy
	{{- end}}
{{- end}}

	return validated, nil
}

{{- range .DomainFields }}
{{- if .NewName | ne ""}}
type {{.NewName}}Valid {{.FieldType}} // Direct copy
{{- else}}
type {{.FieldName}}Valid {{.FieldType}} // Direct copy
{{- end}}
{{- end }}
`
