// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/machinery"
	"log"
	"github.com/seecis/sauron/internal/dataaccess"
	"net/http"
	"time"
	"github.com/spf13/viper"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Spins up a worker sauron to rule on",
	// Todo: expand documentation
	Long: `Workers are capable of running scheduled jobs`,
	Run: func(cmd *cobra.Command, args []string) {
		ew := ExtractionWorker{
			reportService:    dataaccess.NewMSSQLReportService(false, false),
			extractorService: dataaccess.NewMsSqlExtractorService(false, false),
		}

		v := viper.GetString("machinery-broker")
		master := machinery.NewMachineryWithBrokerAddress(v)
		w := master.NewWorker("worker", 1)
		master.RegisterTask("extract", ew.Extract)
		err := w.Launch()
		// todo: maybe add a worker scheduling mechanism?
		// Todo: maybe switch to actors?????
		if err != nil {
			log.Fatal(err)
		}
	},
}

type ExtractionWorker struct {
	reportService    dataaccess.ReportService
	extractorService dataaccess.ExtractorService
}

func (ew *ExtractionWorker) Extract(url, extractorId, reportId string) error {
	extractor, err := ew.extractorService.Get(extractorId)
	if err != nil {
		return err
	}

	hc := http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := hc.Get(url)
	if err != nil {
		// Todo: Figure this out
		return err
	}

	defer res.Body.Close()
	fields, err := extractor.Extract(res.Body)
	if err != nil {
		return err
	}

	err = ew.reportService.WriteAsReport(reportId, fields)
	return err
}

var machineryBrokerAddress string

func init() {
	rootCmd.AddCommand(workerCmd)

	//Todo: maybe add another scheduling mechanism
	workerCmd.Flags().StringVar(&machineryBrokerAddress,
		"machinery-broker",
		"redis://192.168.99.100:6379",
		"Provide address for machinery")

	viper.BindPFlag("machinery-broker", workerCmd.Flags().Lookup("machinery-broker"))
}
