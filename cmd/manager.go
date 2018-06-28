// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/pkg/task-manager"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Sauron's secretary",
	Long:  "Task scheduler. Better version of cron.",
	Run: func(cmd *cobra.Command, args []string) {
		task_manager.Main()
	},
}

var serveManagerCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves json api for scheduling things",
	Run: func(cmd *cobra.Command, args []string) {
		task_manager.Serve(managerIp, managerPort)
	},
}

var managerPort string
var managerIp string

func init() {
	rootCmd.AddCommand(managerCmd)
	managerCmd.AddCommand(serveManagerCmd)
	serveManagerCmd.Flags().StringVarP(&file, "port", "p", "8080", "Port to bind")
	serveManagerCmd.Flags().StringVarP(&file, "ip", "", "0.0.0.0", "Ip to bind")
}
