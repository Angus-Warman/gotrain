package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func addModelCommands(root *cobra.Command) {
	sub := &cobra.Command{
		Use:   "model",
		Short: "Modify shared entities",
	}

	generateCmd := &cobra.Command{
		Use:   "generate [name]",
		Short: "Generate a new model",
		Args:  cobra.ExactArgs(1),
		Run:   generateModel,
	}

	updateCmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update an existing new model",
		Args:  cobra.ExactArgs(1),
		Run:   updateModel,
	}

	sub.AddCommand(
		generateCmd,
		updateCmd,
	)

	root.AddCommand(sub)
}

func generateModel(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Println("Generating model:", name)
}

func updateModel(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Println("Updating model:", name)
}
