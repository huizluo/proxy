package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"proxy/pkg/provider/tunnel"
)

func NewTunnelCmd()*cobra.Command{

	// httpCmd represents the http command
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "run a proxy on tunnel mode",
		Long: `run a proxy on tunnel mode`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return tunnelCmd
}

func NewTunnelClientCmd()*cobra.Command{
	o:=tunnel.TunnelClientArgs{}
	provider:=tunnel.NewTunnelClient()
	// httpCmd represents the http command
	client := &cobra.Command{
		Use:   "client",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			provider.Start(o)
			log.Println("tunnel client closed")
		},
	}

	o.Args = baseArgs
	client.Flags().IntVarP(&o.Timeout,"timeout","t",2000,"tcp timeout with milliseconds")
	client.Flags().BoolVar(&o.IsUDP,"udp",false,"proxy on udp protocol")
	client.Flags().StringVar(&o.Key,"k","default","key same with server")
	return client
}

func NewTunnelServerCmd()*cobra.Command{
	provider:=tunnel.NewTunnelServer()
	o:=tunnel.TunnelServerArgs{}
	// httpCmd represents the http command
	server := &cobra.Command{
		Use:   "server",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			provider.Start(o)
			log.Println("tunnel server closed")
		},
	}

	o.Args = baseArgs
	server.Flags().IntVarP(&o.Timeout,"timeout","t",2000,"tcp timeout with milliseconds")
	server.Flags().BoolVar(&o.IsUDP,"udp",false,"proxy on udp protocol")
	server.Flags().StringVar(&o.Key,"k","default","key same with client")

	return server
}

func NewTunnelBridgeCmd()*cobra.Command{
	provider:=tunnel.NewTunnelBridge()
	o:=tunnel.TunnelBridgeArgs{}
	// httpCmd represents the http command
	bridgeCmd := &cobra.Command{
		Use:   "bridge",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			provider.Start(o)
			fmt.Println("tunnel bridge closed")
		},
	}

	o.Args = baseArgs
	bridgeCmd.Flags().IntVarP(&o.Timeout,"timeout","t",2000,"tcp timeout with milliseconds")

	return bridgeCmd
}
