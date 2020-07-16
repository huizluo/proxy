/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"proxy/pkg/provider"

	"github.com/spf13/cobra"
)



func NewHttpCmd()*cobra.Command{
	var pro = provider.NewHttpProvider()
	o:=provider.HTTPArgs{}
	// httpCmd represents the http command
	httpCmd := &cobra.Command{
		Use:   "http",
		Short: "run a http proxy server",
		Long: `run a http proxy server`,
		Run: func(cmd *cobra.Command, args []string) {
			pro.Start(o)
			fmt.Println("http called")
		},
	}

	o.Args = baseArgs

	httpCmd.Flags().StringVarP(&o.LocalType,"local-type","t","tcp","local type use tls | tcp")
	httpCmd.Flags().StringVarP(&o.ParentType,"parent-type","T","tcp","parent type use tls | tcp")
	httpCmd.Flags().BoolVar(&o.Always,"always",false,"parent type use tls | tcp")
	httpCmd.Flags().IntVar(&o.Timeout,"timeout",2000,"tcp timeout milliseconds when connect to real server or parent proxy")
	httpCmd.Flags().IntVar(&o.HTTPTimeout,"http-timeout",3000,"http request timeout milliseconds when connect to host")
	httpCmd.Flags().IntVar(&o.Interval,"interval",10,"check domain if blocked every interval seconds")
	httpCmd.Flags().StringVarP(&o.Blocked,"blocked","b","blocked","blocked domain file , one domain each line")
	httpCmd.Flags().StringVarP(&o.Direct,"direct","d","direct","direct domain file , one domain each line")
	httpCmd.Flags().StringVarP(&o.AuthFile,"auth-file","F","","http basic auth file,\"username:password\" each line in file")
	httpCmd.Flags().StringArrayVarP(&o.Auth,"auth","a",[]string{},"http basic auth username and password, mutiple user repeat -a ,such as: -a user1:pass1 -a user2:pass2")
	httpCmd.Flags().IntVarP(&o.PoolSize,"pool-size","L",20,"conn pool size , which connect to parent proxy, zero: means turn off pool")
	httpCmd.Flags().IntVarP(&o.CheckParentInterval,"check-parent-interval","I",3,"check if proxy is okay every interval seconds,zero: means no check")

	return httpCmd
}