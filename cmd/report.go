// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/dataaccess"
	"log"
	"fmt"
	"github.com/seecis/sauron/cmd/util"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report tasks",
	Run: func(cmd *cobra.Command, args []string) {
		//if len(args) == 0 {
		//	cmd.Help()
		//	return
		//}
		//
		//id := args[0]
		//rs := dataaccess.NewMSSQLReportService(true, false)
		//report, err := rs.GetReportDetail(id)
		//
		//if err != nil {
		//	log.Fatal(err)
		//	os.Exit(2)
		//	return
		//}
		//
		//r := (&BasicReport{}).from(*report)
		////todo: add format thingie
		//err = yaml.NewEncoder(os.Stdout).Encode(r)
		//if err != nil {
		//	log.Fatal(err)
		//	os.Exit(2)
		//	return
		//}
	},
}

type BasicField struct {
	SubFields []BasicField `yaml:"subFields,omitempty"`
	Label     string       `yaml:"label,omitempty"`
	Data      string       `yaml:"data,omitempty"`
}

func (bf *BasicField) from(field dataaccess.Field) BasicField {
	var subfields []BasicField
	for _, v := range field.SubFields {
		subfields = append(subfields, (&BasicField{}).from(v))
	}

	return BasicField{
		SubFields: subfields,
		Label:     field.Label,
		Data:      field.Data,
	}
}

var reportListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists executed reports",
	Run: func(cmd *cobra.Command, args []string) {
		reports, err := util.DefaultSauronHttpClient.GetReportHeaders()
		if err != nil {
			log.Fatal(err)
		}

		for k, v := range reports {
			fmt.Printf("%d %s %s \n", k+1, v.UID.String(), v.UpdatedAt.String())
		}
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.AddCommand(reportListCmd)
}
