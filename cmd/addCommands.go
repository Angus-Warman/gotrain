package cmd

import (
	"fmt"

	"github.com/Angus-Warman/gotrain/config"
	"github.com/spf13/cobra"
)

func AddCommands(root *cobra.Command) {
	addCreateProjectCommand(root)
	addGenerateCommands(root)
	addRunCommands(root)
}

func requiresAppPath(cmd *cobra.Command, args []string) error {
	if config.AppPath == "" {
		return fmt.Errorf("--project (-p) is required")
	}

	return nil
}

func requiresAppPathAndDBPath(cmd *cobra.Command, args []string) error {
	if config.AppPath == "" {
		return fmt.Errorf("--project (-p) is required")
	}

	if config.DBPath == "" {
		return fmt.Errorf("--db-path (-d) is required")
	}

	return nil
}
