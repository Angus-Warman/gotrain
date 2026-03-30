package db

import (
	"errors"
	"fmt"

	"github.com/Angus-Warman/gotrain/config"
	"github.com/glebarez/sqlite"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func AddCommands(root *cobra.Command) {
	sub := &cobra.Command{
		Use:   "db",
		Short: "Modify database",
	}

	generateCmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a new migration from model states",
		PreRunE: requiresAppPath,
		RunE:    generateMigrationCmd,
	}

	migrateCmd := &cobra.Command{
		Use:     "migrate",
		Short:   "Run all migrations against target DB",
		PreRunE: requiresAppPathAndDBPath,
		RunE:    runMigrationsCmd,
	}

	sub.AddCommand(
		generateCmd,
		migrateCmd,
	)

	root.AddCommand(sub)
}

func requiresAppPath(cmd *cobra.Command, args []string) error {
	if config.AppPath == "" {
		return fmt.Errorf("target --path (-p) is required")
	}

	return nil
}

func requiresAppPathAndDBPath(cmd *cobra.Command, args []string) error {
	if config.AppPath == "" {
		return fmt.Errorf("--path (-p) is required")
	}

	if config.DBPath == "" {
		return fmt.Errorf("--db-path (-d) is required")
	}

	return nil
}

func generateMigrationCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("Generating migrations...")

	if config.AppPath == "" {
		return errors.New("AppPath not set")
	}

	return GenerateMigration(config.AppPath)
}

func runMigrationsCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("Running migrations...")

	if config.AppPath == "" {
		return errors.New("AppPath not set")
	}

	if config.DBPath == "" {
		return errors.New("DBPath not set")
	}

	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{})

	if err != nil {
		return err
	}

	runMigrationsForReal(db, config.AppPath)

	return nil
}
