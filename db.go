package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func addDBCommands(root *cobra.Command) {
	sub := &cobra.Command{
		Use:   "db",
		Short: "Modify database",
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a new migration from model states",
		Run:   generateMigrationCmd,
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run all migrations against target DB",
		Run:   runMigrationsCmd,
	}

	sub.AddCommand(
		generateCmd,
		migrateCmd,
	)

	root.AddCommand(sub)
}

func generateMigrationCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Generating migrations...")
}

func runMigrationsCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Running migrations...")
}
