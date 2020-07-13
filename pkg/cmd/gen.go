package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"proxy/pkg/utils"
)

func NewGenCmd()*cobra.Command{

	// httpCmd represents the http command
	Cmd := &cobra.Command{
		Use:   "gen",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			 if err:=utils.Keygen();err!=nil{
			 	fmt.Printf("gen cert key err,ERR: %s",err.Error())
			 }
		},
	}

	return Cmd
}

func init() {
	rootCmd.AddCommand(NewGenCmd())
}
