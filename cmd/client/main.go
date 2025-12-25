package main

import (
	"service-center/cmd/app/cli"

	"github.com/spf13/cobra"
)

func main() {
	var command *cobra.Command
	cli.Run(command)
}
