package cmd

import (
	"github.com/spf13/cobra"
)

func NewTunnelCmd()*cobra.Command{

	// httpCmd represents the http command
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return tunnelCmd
}

func NewTunnelClientCmd()*cobra.Command{

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

		},
	}

	return client
}

func NewTunnelServerCmd()*cobra.Command{

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

		},
	}

	return server
}

func NewTunnelBridgeCmd()*cobra.Command{

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

		},
	}

	return bridgeCmd
}


func init() {
	tunnelCmd:=NewTunnelCmd()
	rootCmd.AddCommand(tunnelCmd)
	tunnelCmd.AddCommand(NewTunnelClientCmd())
	tunnelCmd.AddCommand(NewTunnelServerCmd())
	tunnelCmd.AddCommand(NewTunnelBridgeCmd())
}
