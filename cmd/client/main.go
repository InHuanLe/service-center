package main

import (
	"service-center/cmd/app/cli"
	"service-center/cmd/app/client"
)

func main() {
	command := client.NewCommand()
	cli.Run(command)
}
