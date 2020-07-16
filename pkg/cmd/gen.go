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
		Short: "gen tls key and crt file",
		Long: `gen tls key and crt file`,
		Run: func(cmd *cobra.Command, args []string) {
			 if err:=utils.Keygen();err!=nil{
			 	fmt.Printf("gen cert key err,ERR: %s",err.Error())
			 }
		},
	}

	return Cmd
}

func init() {
	//rootCmd.AddCommand(NewGenCmd())
	if err:=utils.Keygen();err!=nil{
		panic(err)
	}
}
