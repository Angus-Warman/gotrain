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

//go:embed templates/*.templ
var templateFiles embed.FS

var templateSet = template.Must(template.ParseFS(templateFiles, "templates/*.templ"))

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

	templateName := target + ".templ"

	return writeTemplateForce(templateName, filePath, data)
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

	templateName := target + ".templ"

	return writeTemplateForce(templateName, filePath, data)
}

func WriteTemplateToFolderForce(target, subFolder string, data map[string]any) error {
	appPath := config.AppPath

	if appPath == "" {
		return fmt.Errorf("cannot write template without AppPath")
	}

	filePath := filepath.Join(appPath, subFolder, target)

	templateName := target + ".templ"

	return writeTemplateForce(templateName, filePath, data)
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

	templateName := target // Already has .templ, and should keep it

	return writeTemplateForce(templateName, dstPath, data)
}

func writeTemplateForce(templateName, dstPath string, data map[string]any) error {
	dstFolder := filepath.Dir(dstPath)

	err := os.MkdirAll(dstFolder, 0o755)
	if err != nil {
		return err
	}

	tmpl := templateSet.Lookup(templateName)

	if tmpl == nil {
		return fmt.Errorf("No template with name %s", templateName)
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
