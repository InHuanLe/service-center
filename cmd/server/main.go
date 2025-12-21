/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "service-center/pkg/cli"

func main() {
	command := NewRegistryCommand()
	cli.Run(command)
}
