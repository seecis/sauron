// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/machinery"
	"github.com/pborman/uuid"
	"log"
)

var extractionCmd = &cobra.Command{
	Use:   "extraction",
	Short: "Schedules an extraction job",
	Long:  `Schedules an extraction job for later execution. If a report id is not specified it will be created`,
	Run: func(cmd *cobra.Command, args []string) {
		m := machinery.NewMachinery()
		_, err := m.SendTask(machinery.NewExtractionJob(url, extractorId, reportID))
		if err != nil {
			log.Fatal(err)
		}

		println(reportID)
	},
}

var url string
var extractorId string
var reportID string

func init() {
	scheduleCmd.AddCommand(extractionCmd)

	extractionCmd.Flags().StringVar(&url, "url", "", "URL that extractor will be ran on")
	extractionCmd.Flags().StringVarP(&extractorId, "extractor", "e", "", "extractor name")
	extractionCmd.Flags().StringVarP(&reportID, "reportId", "r", uuid.New(), "reportId name")

	extractionCmd.MarkFlagRequired("url")
	extractionCmd.MarkFlagRequired("extractor")
}
