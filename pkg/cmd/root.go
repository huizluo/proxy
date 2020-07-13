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
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"proxy/pkg/provider"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile,certPath,keyPath string
var baseArgs provider.Args

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "this is a network proxy",
	Long: `proxy is a goland dev proxy tool. it support http tcp socket5 udp https proxy etc.`,
	Annotations: map[string]string{"author":"huizluo"},
	Version: "1.0",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	baseArgs = provider.Args{}

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.proxy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().StringVarP(&baseArgs.Parent,"parent","P","","parent address, such as: \"23.32.32.19:28008\"")
	rootCmd.Flags().StringVarP(&baseArgs.Local,"local","p",":30080","local listen addr ip:port")
	rootCmd.Flags().StringVarP(&certPath,"cert","C","proxy.crt","tls cert file")
	rootCmd.Flags().StringVarP(&keyPath,"key","K","proxy.key","tls private key file")
	//crtBytes,keyBytes:=readTlsKeyFile()
	//baseArgs.CertBytes = crtBytes
	//baseArgs.KeyBytes = keyBytes
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".proxy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".proxy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func readTlsKeyFile()(crtbytes,keyBytes []byte){
	crtbytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		log.Fatalf("err : %s", err)
		return
	}
	keyBytes, err = ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("err : %s", err)
		return
	}
	return
}
