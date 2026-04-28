package cmd

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
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

	slog.Info("Created model: " + modelName)
	return nil
}

var bannedModelNames = []string{
	"ext",
	"gen",
	"public",
}

func createModel(appPath, modelName string, propertyStrings []string) error {
	slog.Debug("Creating model: " + modelName)

	if appPath == "" || modelName == "" {
		return fmt.Errorf("Missing parameters")
	}

	if len(modelName) < 2 {
		return fmt.Errorf("Model name must be at least two characters")
	}

	modelName = ext.CapitaliseFirst(modelName)

	modelLower := strings.ToLower(modelName)

	if slices.Contains(bannedModelNames, modelLower) {
		return fmt.Errorf("Cannot create model with name %v", modelLower)
	}

	modelFolder := filepath.Join(appPath, modelLower)

	err := os.MkdirAll(modelFolder, 0o755)

	if err != nil {
		return err
	}

	modelProperties, err := modelPropertiesFromStrings(propertyStrings)

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

	return nil
}

func generateMigrationCmd(cmd *cobra.Command, args []string) error {
	err := database.GenerateMigration(config.AppPath)

	if err != nil {
		return err
	}

	slog.Info("Generated migrations")
	return nil
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

	slog.Info("Generated boilerplate for: " + modelName)
	return nil
}

func generateBoilerplateForModel(appPath, modelName string) error {
	slog.Debug("Generating boilerplate for: " + modelName)

	modelName = ext.CapitaliseFirst(modelName)
	modelLower := strings.ToLower(modelName)

	modelProperties, err := getModelProperties(appPath, modelName)

	if err != nil {
		return err
	}

	navLinks, err := navBarLinks(appPath)

	if err != nil {
		return err
	}

	data := map[string]any{
		"projectName": ext.ProjectName(appPath),
		"ModelName":   modelName,
		"modelName":   modelLower,
		"Properties":  modelProperties,
		"NavLinks":    navLinks,
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

	err = updateSharedFiles(appPath)

	if err != nil {
		return err
	}

	return nil
}

func updateSharedFiles(appPath string) error {
	err := updateAddHandlers(appPath)

	if err != nil {
		return err
	}

	err = updateIndexHtml(appPath)

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

type Link struct {
	Href  string
	Label string
}

func navBarLinks(appPath string) ([]Link, error) {
	modelNames, err := findModelNames(appPath)

	if err != nil {
		return nil, err
	}

	links := []Link{
		{
			Href:  "/",
			Label: "Home",
		},
	}

	for _, modelName := range modelNames {
		link := Link{
			Label: ext.CapitaliseFirst(modelName),
			Href:  fmt.Sprintf("/%v", modelName),
		}

		links = append(links, link)
	}

	return links, nil
}

func updateAddHandlers(appPath string) error {
	slog.Debug("Updating AddHandlers.go")

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

func updateIndexHtml(appPath string) error {
	slog.Debug("Updating index.html")

	links, err := navBarLinks(appPath)

	if err != nil {
		return err
	}

	indexHtmlPath := filepath.Join(appPath, "public", "index.html")

	currentBytes, err := os.ReadFile(indexHtmlPath)

	// If file already exists, check for flag comment
	if err == nil {
		currentHTML := string(currentBytes)
		flagComment := "<!-- This file will be updated whenever a new model is added, unless this comment is removed -->"

		if !strings.HasPrefix(currentHTML, flagComment) {
			slog.Debug("Leaving index.html unchanged, flag comment not found")
			return nil
		}
	}

	data := map[string]any{
		"projectName": ext.ProjectName(appPath),
		"NavLinks":    links,
	}

	err = filing.WriteTemplateToFolderForce("index.html", "public", data)

	if err != nil {
		return err
	}

	return nil
}
