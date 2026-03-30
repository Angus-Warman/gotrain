package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gotrain",
		Short: "gotrain: go boilerplate generator",
	}

	addModelCommands(rootCmd)
	addDBCommands(rootCmd)

	err := rootCmd.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
