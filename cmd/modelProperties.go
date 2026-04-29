package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Angus-Warman/gotrain/ext"
	"github.com/Angus-Warman/gotrain/parse"
)

type ModelProperty struct {
	Name string
	Type string
	Tag  string
}

func createModelProperty(propertyString string) (ModelProperty, error) {
	parts := strings.Split(propertyString, ":")

	name := parts[0]
	name = ext.CapitaliseFirst(name)

	typeString := "string" // Default

	tags := []string{}

	for _, part := range parts[1:] {
		if value, ok := strings.CutPrefix(part, "default="); ok {
			tags = append(tags, "default:"+value)
			continue
		}

		switch part {
		case "required":
			tags = append(tags, "not null")

		case "unique":
			tags = append(tags, "unique")

		case "int":
			fallthrough
		case "integer":
			typeString = "int"

		case "number":
			fallthrough
		case "real":
			fallthrough
		case "float":
			typeString = "float32"

		case "bool":
			fallthrough
		case "boolean":
			fallthrough
		case "yes/no":
			fallthrough
		case "y/n":
			fallthrough
		case "true/false":
			typeString = "bool"
		}
	}

	tag := ""

	if len(tags) > 0 {
		gormTags := strings.Join(tags, ";")
		tag = fmt.Sprintf("`gorm:\"%v\"`", gormTags)
	}

	property := ModelProperty{
		Name: name,
		Type: typeString,
		Tag:  tag,
	}

	return property, nil
}

func modelPropertiesFromStrings(propertyStrings []string) ([]ModelProperty, error) {
	properties := make([]ModelProperty, len(propertyStrings))

	for i, propertyString := range propertyStrings {
		property, err := createModelProperty(propertyString)

		if err != nil {
			return nil, err
		}

		properties[i] = property
	}

	return properties, nil
}

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
}

type Property struct {
	Name     string
	Type     string
	HTMLType string
}

func getModelProperties(appPath, modelName string) ([]Property, error) {
	modelName = strings.ToLower(modelName)

	filePath := filepath.Join(appPath, modelName, "model.go")
	fields, err := parse.ParseModel(filePath)

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

		fieldName := strings.ToLower(field.Name)

		if strings.Contains(fieldName, "email") {
			htmlType = "email"
		}

		if strings.Contains(fieldName, "phone") {
			htmlType = "tel"
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
