package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Angus-Warman/gotrain/cmd"
	"github.com/Angus-Warman/gotrain/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gotrain",
		Short: "gotrain: go boilerplate generator",

		PersistentPreRunE: handleFlags,
	}

	addFlags(rootCmd)

	cmd.AddCommands(rootCmd)

	err := rootCmd.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addFlags(root *cobra.Command) {
	root.PersistentFlags().StringVarP(
		&config.AppPath,
		"project",
		"p",
		"",
		"target project path",
	)

	root.PersistentFlags().StringVarP(
		&config.DBPath,
		"db-path",
		"d",
		"",
		"target database path",
	)

	root.PersistentFlags().BoolVarP(
		&config.Force,
		"force",
		"f",
		false,
		"overwrite existing files",
	)
}

func handleFlags(cmd *cobra.Command, args []string) error {
	projectPathSet := cmd.Flags().Changed("project")

	if !projectPathSet {
		cwd, err := os.Getwd()

		if err != nil {
			return fmt.Errorf("error getting working directory: %w", err)
		}

		if filepath.Base(cwd) == "gotrain" {
			return fmt.Errorf("cannot run from within the gotrain source directory")
		}

		config.AppPath = cwd
	}

	absAppPath, err := filepath.Abs(config.AppPath)

	if err != nil {
		return fmt.Errorf("error resolving app path: %w", err)
	}

	config.AppPath = absAppPath

	dbPathSet := cmd.Flags().Changed("db-path")

	if !dbPathSet {
		config.DBPath = filepath.Join(config.AppPath, "data", "data.db")
	} else {
		absDBPath, err := filepath.Abs(config.DBPath)

		if err != nil {
			return fmt.Errorf("error resolving db path: %w", err)
		}

		config.DBPath = absDBPath
	}

	return nil
}
