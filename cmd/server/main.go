package main

import (
	"service-center/cmd/app/cli"
	cmd "service-center/cmd/app/server"
)

func main() {
	command := cmd.NewCommand()
	cli.Run(command)
}
