package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Angus-Warman/gotrain/config"
	"github.com/Angus-Warman/gotrain/database"
	"github.com/Angus-Warman/gotrain/ext"
	"github.com/Angus-Warman/gotrain/filing"
	"github.com/spf13/cobra"
)

func addGenerateCommands(root *cobra.Command) {
	sub := &cobra.Command{
		Use:   "generate",
		Short: "Create models, handlers and migrations",
	}

	sub.AddCommand(
		&cobra.Command{
			Use:     "model [name] [properties...]",
			Short:   "Create a new model",
			Long:    "Model name should be singular, not plural. Model properties are constructed like so, 'DisplayName', 'Email:required:unique', 'Quantity:int', 'Price:float'",
			Args:    cobra.MinimumNArgs(1),
			PreRunE: requiresAppPath,
			RunE:    createModelCmd,
		},

		&cobra.Command{
			Use:     "handlers [model name]",
			Short:   "HTTP endpoints and templates",
			Args:    cobra.ExactArgs(1),
			PreRunE: requiresAppPath,
			RunE:    generateBoilerplateCmd,
		},

		&cobra.Command{
			Use:     "migration",
			Short:   "DB migration for all model changes",
			PreRunE: requiresAppPathAndDBPath,
			RunE:    generateMigrationCmd,
		},
	)

	root.AddCommand(sub)
}

func createModelCmd(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	properties := args[1:]

	err := createModel(config.AppPath, modelName, properties)

	if err != nil {
		return err
	}

	fmt.Println("Model created")
	return nil
}

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
		switch part {
		case "required":
			tags = append(tags, "not null")

		case "unique":
			tags = append(tags, "unique")

		case "int":
			typeString = "int"

		case "number":
			typeString = "float32"

		case "float":
			typeString = "float32"
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

func createModelProperties(propertyStrings []string) ([]ModelProperty, error) {
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

func createModel(appPath, modelName string, propertyStrings []string) error {
	if appPath == "" || modelName == "" {
		return fmt.Errorf("Missing parameters")
	}

	if len(modelName) < 2 {
		return fmt.Errorf("Model name must be at least two characters")
	}

	modelName = ext.CapitaliseFirst(modelName)

	modelLower := strings.ToLower(modelName)

	modelFolder := filepath.Join(appPath, modelLower)

	err := os.MkdirAll(modelFolder, 0o755)

	if err != nil {
		return err
	}

	modelProperties, err := createModelProperties(propertyStrings)

	if err != nil {
		return err
	}

	modelTarget := "model.go"

	data := map[string]any{
		"ModelName":  modelName,
		"modelName":  modelLower,
		"properties": modelProperties,
	}

	err = filing.WriteTemplateToFolder(modelTarget, modelLower, data)

	if err != nil {
		return err
	}

	err = generateBoilerplateForModel(appPath, modelName)

	if err != nil {
		return err
	}

	err = updateAddHandlers(appPath)

	if err != nil {
		return err
	}

	return nil
}

func generateMigrationCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating migrations...")

	return database.GenerateMigration(config.AppPath)
}

func generateBoilerplateCmd(cmd *cobra.Command, args []string) error {
	modelName := args[0]

	if modelName == "" {
		return fmt.Errorf("model cannot be empty")
	}

	appPath := config.AppPath

	err := generateBoilerplateForModel(appPath, modelName)

	if err != nil {
		return err
	}

	fmt.Println("Boilerplate generated")
	return nil
}

func generateBoilerplateForModel(appPath, modelName string) error {
	modelName = ext.CapitaliseFirst(modelName)
	modelLower := strings.ToLower(modelName)

	modelProperties, err := getModelProperties(appPath, modelName)

	if err != nil {
		return err
	}

	data := map[string]any{
		"projectName": ext.ProjectName(appPath),
		"ModelName":   modelName,
		"modelName":   modelLower,
		"Properties":  modelProperties,
	}

	standardTemplates := []string{
		"handler.go",
		"store.go",
		"page.html",
	}

	for _, target := range standardTemplates {
		err := filing.WriteTemplateToFolder(target, modelLower, data)

		if err != nil {
			return err
		}
	}

	metaTemplates := []string{
		"table.html.templ",
		"editDialog.html.templ",
		"addDialog.html.templ",
	}

	for _, target := range metaTemplates {
		err := filing.WriteMetaTemplateToFolder(target, modelLower, data)

		if err != nil {
			return err
		}
	}

	err = updateAddHandlers(appPath)

	if err != nil {
		return err
	}

	return nil
}

func findModelNames(appPath string) ([]string, error) {
	var modelNames []string

	err := filepath.WalkDir(appPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && d.Name() == "model.go" {
			modelName := filepath.Base(filepath.Dir(path))
			modelNames = append(modelNames, modelName)
		}

		return nil
	})

	return modelNames, err
}

func updateAddHandlers(appPath string) error {
	projectName := ext.ProjectName(appPath)
	modelNames, err := findModelNames(appPath)

	if err != nil {
		return err
	}

	data := map[string]any{
		"projectName": projectName,
		"modelNames":  modelNames,
	}

	target := "AddHandlers.go"

	err = filing.WriteTemplateToFolderForce(target, "gen", data)

	if err != nil {
		return err
	}

	return nil
}
