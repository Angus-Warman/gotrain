package filing

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Angus-Warman/gotrain/config"
)

//go:embed templates
var templatesFS embed.FS

func WriteTemplate(target string, data map[string]any) error {
	appPath := config.AppPath

	if appPath == "" {
		return fmt.Errorf("cannot write template without AppPath")
	}

	filePath := filepath.Join(appPath, target)

	if fileExists(filePath) {
		if !config.Force {
			return fmt.Errorf("file already exists: %s", filePath)
		}
	}

	templatePath := "templates/" + target + ".templ"

	return writeTemplateForce(filePath, templatePath, data)
}

func WriteTemplateToFolder(target, subFolder string, data map[string]any) error {
	appPath := config.AppPath

	if appPath == "" {
		return fmt.Errorf("cannot write template without AppPath")
	}

	filePath := filepath.Join(appPath, subFolder, target)

	if fileExists(filePath) {
		if !config.Force {
			return fmt.Errorf("file already exists: %s", filePath)
		}
	}

	templatePath := "templates/" + target + ".templ"

	return writeTemplateForce(filePath, templatePath, data)
}

func WriteTemplateToFolderForce(target, subFolder string, data map[string]any) error {
	appPath := config.AppPath

	if appPath == "" {
		return fmt.Errorf("cannot write template without AppPath")
	}

	filePath := filepath.Join(appPath, subFolder, target)

	templatePath := "templates/" + target + ".templ"

	return writeTemplateForce(filePath, templatePath, data)
}

func WriteMetaTemplateToFolder(target, subFolder string, data map[string]any) error {
	appPath := config.AppPath

	if appPath == "" {
		return fmt.Errorf("cannot write template without AppPath")
	}

	dstPath := filepath.Join(appPath, subFolder, target)

	if fileExists(dstPath) {
		if !config.Force {
			return fmt.Errorf("file already exists: %s", dstPath)
		}
	}

	templatePath := "templates/" + target // Already has .templ, and should keep it

	return writeTemplateForce(dstPath, templatePath, data)
}

func writeTemplateForce(dstPath, templatePath string, data map[string]any) error {
	dstFolder := filepath.Dir(dstPath)

	err := os.MkdirAll(dstFolder, 0o755)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFS(templatesFS, templatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

//go:embed all:static
var staticFS embed.FS // "all:" includes files that start with "."

func CopyStaticToFolder(target, appPath, folder string) error {
	src := "static/" + target
	target = strings.TrimSuffix(target, ".static")
	dst := filepath.Join(appPath, folder, target)

	data, err := staticFS.ReadFile(src)

	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(dst), 0o755)

	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0o644)
}
