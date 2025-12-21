/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"service-center/pkg/api/pb"
	"service-center/pkg/client/registry"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "registry-client",
		Short: "Service registry client CLI tool",
		Long:  `Command line tool to interact with the service registry server.`,
	}

	// Register command
	registerCmd := &cobra.Command{
		Use:   "register <service-name> <address> [port]",
		Short: "Register a service instance",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverAddr, _ := cmd.Flags().GetString("server")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			client, err := registry.NewClient(serverAddr)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			serviceName := args[0]
			instanceAddr := args[1]
			var port int32 = 0
			if len(args) > 2 {
				var p int
				fmt.Sscanf(args[2], "%d", &p)
				port = int32(p)
			}

			// Generate a simple instance ID (in production, use UUID)
			instanceID := fmt.Sprintf("%s-%d", serviceName, time.Now().Unix())

			success, err := client.Register(ctx, &pb.ServiceInstance{
				ServiceName: serviceName,
				InstanceId:  instanceID,
				Address:     instanceAddr,
				Port:        port,
			})
			if err != nil {
				return fmt.Errorf("failed to register: %w", err)
			}

			if success {
				fmt.Printf("Successfully registered: %s (ID: %s)\n", serviceName, instanceID)
			} else {
				fmt.Printf("Registration failed for: %s\n", serviceName)
			}
			return nil
		},
	}

	// Discover command
	discoverCmd := &cobra.Command{
		Use:   "discover <service-name>",
		Short: "Discover service instances",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverAddr, _ := cmd.Flags().GetString("server")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			client, err := registry.NewClient(serverAddr)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			serviceName := args[0]

			service, err := client.Discover(ctx, serviceName)
			if err != nil {
				return fmt.Errorf("failed to discover: %w", err)
			}

			if len(service.Instances) == 0 {
				fmt.Printf("No instances found for service: %s\n", serviceName)
				return nil
			}

			fmt.Printf("Found %d instance(s) for service %s:\n", len(service.Instances), serviceName)
			for _, inst := range service.Instances {
				if inst.Port > 0 {
					fmt.Printf("  - ID: %s, Address: %s:%d\n", inst.InstanceId, inst.Address, inst.Port)
				} else {
					fmt.Printf("  - ID: %s, Address: %s\n", inst.InstanceId, inst.Address)
				}
			}
			return nil
		},
	}

	// Unregister command
	unregisterCmd := &cobra.Command{
		Use:   "unregister <service-name> <instance-id>",
		Short: "Unregister a service instance",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverAddr, _ := cmd.Flags().GetString("server")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			client, err := registry.NewClient(serverAddr)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			defer client.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			serviceName := args[0]
			instanceID := args[1]

			success, err := client.Deregister(ctx, serviceName, instanceID)
			if err != nil {
				return fmt.Errorf("failed to unregister: %w", err)
			}

			if success {
				fmt.Printf("Successfully unregistered instance: %s\n", instanceID)
			} else {
				fmt.Printf("Failed to unregister instance: %s\n", instanceID)
			}
			return nil
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().String("server", "localhost:50051", "Registry server address")
	rootCmd.PersistentFlags().Duration("timeout", 10*time.Second, "Request timeout")

	// Add subcommands
	rootCmd.AddCommand(registerCmd, discoverCmd, unregisterCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
