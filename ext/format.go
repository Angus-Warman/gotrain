package ext

import (
	"path/filepath"
	"strings"
)

func CapitaliseFirst(str string) string {
	if str == "" {
		return ""
	}

	return strings.ToUpper(string(str[0])) + str[1:]
}

func ProjectName(appPath string) string {
	folderName := filepath.Base(appPath)
	projectName := strings.ToLower(folderName) // Probably!

	return projectName
}
