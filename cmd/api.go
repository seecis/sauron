// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/http"
	"github.com/spf13/viper"
	"github.com/opentracing/opentracing-go/log"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Serves sauron api",
	PreRun: func(cmd *cobra.Command, args []string) {
		p := cmd.Flags().Lookup("port")
		err := viper.BindPFlag("serve.api.port", p)
		if err != nil {
			log.Error(err)
		}

		ip := cmd.Flags().Lookup("ip")
		err = viper.BindPFlag("serve.api.ip", ip)
		if err != nil {
			log.Error(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetString("serve.api.port")
		ip := viper.GetString("serve.api.ip")
		http.ServeApi(ip, port)
	},
}

var backendPort string
var backendIp string

func init() {
	serveCmd.AddCommand(apiCmd)
	apiFlags := apiCmd.Flags()
	apiFlags.StringVarP(&backendPort, "port", "p", "9091", "Port to bind for frontend")
	apiFlags.StringVarP(&backendIp, "ip", "i", "127.0.0.1", "Ip will listen to this ip")
}
