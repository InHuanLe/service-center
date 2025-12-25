package server

import (
	"net"
	pb "service-center/pkg/registry"
	"service-center/pkg/service/registry"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func Run(o *Options) error {
	grpcServer := grpc.NewServer()
	registryService, err := registry.NewRegistryService(o)
	if err != nil {
		return err
	}
	pb.RegisterRegistryServer(grpcServer, registryService)
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   o.IP(),
		Port: o.Port(),
		Zone: o.Zone(),
	})
	if err != nil {
		return err
	}
	return grpcServer.Serve(listener)
}

func NewCommand() *cobra.Command {
	opts := NewOptions()
	command := &cobra.Command{
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
			if err := Run(opts); err != nil {
				return err
			}
			return nil
		},
	}
	opts.AddFlags(command)
	return command
}
