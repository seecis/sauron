// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"fmt"
	"github.com/seecis/sauron/cmd/util"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists available executor ids",
	Run: func(cmd *cobra.Command, args []string) {
		extractors, err := util.DefaultSauronHttpClient.GetExtractors()
		if err != nil {
			log.Fatal(err)
		}

		for k, v := range extractors {
			fmt.Printf("%d %s [%s]\n", k+1, v.GetUid().String(), v.GetName())
		}
	},
}

func init() {
	extractorCmd.AddCommand(listCmd)
}
