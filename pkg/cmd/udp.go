package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"proxy/pkg/provider"
)

func NewUdpCmd()*cobra.Command{
	var pro = provider.NewUdpProvider()
	o:=provider.UDPArgs{}
	// httpCmd represents the http command
	tcpCmd := &cobra.Command{
		Use:   "udp",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err:=pro.Start(o);err!=nil{
				fmt.Errorf("run http proxy fail,err: %s",err.Error())
				return
			}
			fmt.Println("http called")
		},
	}

	o.Args = baseArgs

	tcpCmd.Flags().StringVarP(&o.ParentType,"parent-type","T","tcp","parent type use tls | tcp")
	tcpCmd.Flags().IntVar(&o.Timeout,"timeout",2000,"tcp timeout milliseconds when connect to real server or parent proxy")
	tcpCmd.Flags().IntVarP(&o.PoolSize,"pool-size","L",20,"conn pool size , which connect to parent proxy, zero: means turn off pool")
	tcpCmd.Flags().IntVarP(&o.CheckParentInterval,"check-parent-interval","I",3,"check if proxy is okay every interval seconds,zero: means no check")

	return tcpCmd
}

func init() {
	rootCmd.AddCommand(NewUdpCmd())
}
