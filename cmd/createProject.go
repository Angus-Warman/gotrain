package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Angus-Warman/gotrain/config"
	"github.com/Angus-Warman/gotrain/ext"
	"github.com/Angus-Warman/gotrain/filing"
	"github.com/spf13/cobra"
)

func addCreateProjectCommand(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create new project",
		PreRunE: requiresAppPath,
		RunE:    createNewProjectCmd,
	}

	root.AddCommand(cmd)
}

func folderIsEmpty(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		return false
	}

	defer f.Close()

	_, err = f.Readdirnames(1)

	return err == io.EOF
}

func clearFolder(path string) error {
	slog.Debug("Clearing folder")

	entries, err := os.ReadDir(path)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		err = os.RemoveAll(filepath.Join(path, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func createNewProjectCmd(cmd *cobra.Command, args []string) error {
	slog.Debug("Creating new project")

	appPath := config.AppPath

	if !folderIsEmpty(appPath) {
		if !config.Force {
			return fmt.Errorf("project already exists: %s", appPath)
		}

		// Clear existing
		err := clearFolder(appPath)

		if err != nil {
			return err
		}
	}

	err := copyDefaultFiles(appPath)

	if err != nil {
		return err
	}

	err = generateMain(appPath)

	if err != nil {
		return err
	}

	err = updateSharedFiles(appPath)

	if err != nil {
		return err
	}

	err = ext.GoModuleSetup(appPath)

	if err != nil {
		return err
	}

	err = ext.GoFmt(appPath)

	if err != nil {
		return err
	}

	err = ext.GitRepoSetup(appPath)

	if err != nil {
		return err
	}

	slog.Info("Project created")
	return nil
}

func copyDefaultFiles(appPath string) error {
	slog.Debug("Copying default files")

	publicFiles := []string{
		"styles.css",
		"layout.css",
		"htmx-2.0.8.js",
		"hx-ext-json-enc-2.0.1.js",
		"theme/minimal-modern.css", // TODO: Allow selecting themes
	}

	for _, file := range publicFiles {
		err := filing.CopyStaticToFolder(file, appPath, "public")

		if err != nil {
			return err
		}
	}

	extFiles := []string{
		"httpExt.go.static",
	}

	for _, file := range extFiles {
		err := filing.CopyStaticToFolder(file, appPath, "ext")

		if err != nil {
			return err
		}
	}

	projectFiles := []string{
		"devTools.go.static",
		".gitignore.static",
	}

	for _, file := range projectFiles {
		err := filing.CopyStaticToFolder(file, appPath, "")

		if err != nil {
			return err
		}
	}

	return nil
}

func generateIndexHtml(appPath string) error {
	slog.Debug("Generating index.html")

	return nil
}

func generateMain(appPath string) error {
	slog.Debug("Generating main.go")

	target := "main.go"

	data := map[string]any{
		"projectName": ext.ProjectName(appPath),
	}

	err := filing.WriteTemplate(target, data)

	if err != nil {
		return err
	}

	return nil
}
