// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/cmd/util"
	"fmt"
	"os"
)

var extractionCmd = &cobra.Command{
	Use:   "extraction",
	Short: "Schedules an extraction job",
	Long:  `Schedules an extraction job for later execution. If a report id is not specified it will be created`,
	Run: func(cmd *cobra.Command, args []string) {
		id, err := util.DefaultSauronHttpClient.ScheduleExtraction(extractorId, url)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		fmt.Println(id)
	},
}

var url string
var extractorId string

func init() {
	scheduleCmd.AddCommand(extractionCmd)
	extractionCmd.Flags().StringVar(&url, "url", "", "URL that extractor will be ran on")
	extractionCmd.Flags().StringVarP(&extractorId, "extractor", "e", "", "extractor id")

	extractionCmd.MarkFlagRequired("url")
	extractionCmd.MarkFlagRequired("extractor")
}
