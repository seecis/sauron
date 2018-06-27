// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package machinery

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"log"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/spf13/viper"
)

func NewMachineryWithBrokerAddress(address string) *machinery.Server {
	cnf := config.Config{
		Broker: address,
		//Todo: maybe add this to config too?
		DefaultQueue:  "sauron-html",
		ResultBackend: address,
	}

	srv, err := machinery.NewServer(&cnf)
	if err != nil {
		log.Fatal(err)
	}

	return srv
}

func NewMachinery() *machinery.Server {
	return NewMachineryWithBrokerAddress(viper.GetString("machinery-broker"))
}

func NewWorker() {
	m := NewMachinery()
	w := m.NewWorker("worker-1", 2)
	w.Launch()
}

func NewExtractionJob(url string, extractorId, reportId string) *tasks.Signature {
	sig := tasks.Signature{
		Name: "extract",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: url,
			}, {
				Type:  "string",
				Value: extractorId,
			}, {
				Type:  "string",
				Value: reportId,
			},
		},
	}

	return &sig
}
