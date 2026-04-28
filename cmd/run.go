package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Angus-Warman/gotrain/config"
	"github.com/Angus-Warman/gotrain/database"
	"github.com/glebarez/sqlite"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func addRunCommands(root *cobra.Command) {
	sub := &cobra.Command{
		Use:   "run",
		Short: "Run scripts",
	}

	sub.AddCommand(
		&cobra.Command{
			Use:     "migrations",
			PreRunE: requiresAppPathAndDBPath,
			RunE:    runMigrationsCmd,
		},
	)

	root.AddCommand(sub)
}

func runMigrationsCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("Running migrations...")

	folder := filepath.Dir(config.DBPath)

	err := os.MkdirAll(folder, 0o755)

	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{})

	if err != nil {
		return err
	}

	database.RunMigrations(db, config.AppPath)

	return nil
}
