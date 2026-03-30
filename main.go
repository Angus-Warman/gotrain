package main

import (
	"fmt"
	"os"

	"github.com/Angus-Warman/gotrain/config"
	"github.com/Angus-Warman/gotrain/db"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gotrain",
		Short: "gotrain: go boilerplate generator",
	}

	addFlags(rootCmd)

	db.AddCommands(rootCmd)

	err := rootCmd.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addFlags(root *cobra.Command) {
	root.PersistentFlags().StringVarP(
		&config.AppPath,
		"path",
		"p",
		"",
		"target app folder path",
	)

	root.PersistentFlags().StringVarP(
		&config.DBPath,
		"db-path",
		"d",
		"./data.db",
		"target database path",
	)
}
