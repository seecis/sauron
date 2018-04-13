// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package machinery

import (
	"testing"
	"github.com/davecgh/go-spew/spew"
	"github.com/RichardKnop/machinery/v1/tasks"
	"fmt"
	"time"
)

func TestNewMachinery(t *testing.T) {
	m := NewMachinery()

	sig := tasks.Signature{
		Name: "add",

		Args: []tasks.Arg{{
			Type:  "int64",
			Value: 2,
		}, {
			Type:  "int64",
			Value: 41414123,
		},
		},
	}

	asyncResult, err := m.SendTask(&sig)
	if err != nil {
		fmt.Printf("Could not send task: %s", err.Error())
		t.FailNow()
	}

	res, err := asyncResult.Get(time.Duration(time.Millisecond) * 5)
	if err != nil {
		fmt.Printf("Could not send task: %s", err.Error())
	}
	spew.Dump(res)
}


func TestNewWorkerWorkerTest(t *testing.T) {
	NewWorker()
}