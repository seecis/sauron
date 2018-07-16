// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/pkg/task-manager"
	"github.com/spf13/viper"
	"github.com/jinzhu/gorm"
	"fmt"
	"log"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Sauron's secretary",
	Long:  "Task scheduler. Better version of cron.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var serveManagerCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves json api for scheduling things",
	Run: func(cmd *cobra.Command, args []string) {

		dbConf := viper.GetStringMap("manager.database")
		u := dbConf["username"]
		p := dbConf["password"]
		h := dbConf["host"]
		po := dbConf["port"]
		dbname := dbConf["database"]

		g, err := gorm.Open("mssql", fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", u, p, h, po, dbname))
		if err != nil {
			log.Fatal(err)
		}

		ip := viper.GetString("manager.ip")
		port := viper.GetString("manager.port")
		task_manager.Serve(ip, port, g)
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
