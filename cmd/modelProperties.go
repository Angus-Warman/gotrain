package cmd

import (
	"path/filepath"
	"strings"

	"github.com/Angus-Warman/gotrain/ext"
)

var htmlTypeMap = map[string]string{
	"string":    "text",
	"bool":      "checkbox",
	"int":       "number",
	"int8":      "number",
	"int16":     "number",
	"int32":     "number",
	"int64":     "number",
	"uint":      "number",
	"uint8":     "number",
	"uint16":    "number",
	"uint32":    "number",
	"uint64":    "number",
	"float32":   "number",
	"float64":   "number",
	"time.Time": "date",
	// "[]byte":     reflect.TypeFor[[]byte](),
}

type Property struct {
	Name     string
	Type     string
	HTMLType string
}

func getModelProperties(appPath, modelName string) ([]Property, error) {
	modelName = strings.ToLower(modelName)

	filePath := filepath.Join(appPath, modelName, "model.go")
	fields, err := ext.ParseModel(filePath)

	if err != nil {
		return nil, err
	}

	properties := []Property{}

	for _, field := range fields {
		if field.Name == "ID" {
			continue
		}

		htmlType, ok := htmlTypeMap[field.Type]

		if !ok {
			htmlType = "text"
		}

		if strings.ToLower(field.Name) == "email" {
			htmlType = "email"
		}

		property := Property{
			Name:     field.Name,
			Type:     field.Type,
			HTMLType: htmlType,
		}

		properties = append(properties, property)
	}

	return properties, nil
}
