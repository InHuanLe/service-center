package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Run(command *cobra.Command) {
	if err := command.Execute(); err != nil {
		fmt.Printf("command exit with err: %+v", err)
		os.Exit(1)
	}
}
