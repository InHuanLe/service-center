/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"

	"github.com/spf13/cobra"
)

func NewRegistryCommand() *cobra.Command {
	opts := NewRegistryOptions()
	rootCmd := &cobra.Command{
		Use:   "service-center",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: use background context temporarily, improved later
			if err := opts.Run(context.Background()); err != nil {
				return err
			}
			return nil
		},
	}
	opts.AddFlags(rootCmd)
	return rootCmd
}
