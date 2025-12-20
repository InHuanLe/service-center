/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"service-center/cmd"
	"service-center/pkg/cli"
)

func main() {
	command := cmd.NewRegistryCommand()
	cli.Run(command)
}
