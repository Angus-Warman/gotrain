package ext

import (
	"fmt"
	"log/slog"
	"os/exec"
)

func runCommand(appPath, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = appPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec failed: %w\n%s", err, string(output))
	}

	return nil
}

func GoFmt(appPath string) error {
	return runCommand(appPath, "go", "fmt", "./...")
}

func GoModuleSetup(appPath string) error {
	slog.Debug("Creating go module")

	projectName := ProjectName(appPath)

	err := runCommand(appPath, "go", "mod", "init", projectName)

	if err != nil {
		return err
	}

	err = runCommand(appPath, "go", "mod", "tidy")

	if err != nil {
		return err
	}

	return nil
}

func GitRepoSetup(appPath string) error {
	slog.Debug("Creating git repo")

	err := runCommand(appPath, "git", "init")

	if err != nil {
		return err
	}

	err = runCommand(appPath, "git", "add", ".")

	if err != nil {
		return err
	}

	err = runCommand(appPath, "git", "commit", "-m", "Initial commit")

	if err != nil {
		return err
	}

	return nil
}
