// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"testing"
	"fmt"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
)

func TestTask_Next5Times(t *testing.T) {

	tt := Task{
		TriggerAddress: "http://test.example.com",
		TriggerMethod:  "POST",
		TriggerParams: map[string]interface{}{
			"aa": "bb",
		},
		ErrorCallbackAddress: "http://error.example.com",
		ErrorCallbackMethod:  "POST",
		Timeout:              40,
		Retry: Retry{
			RetryType:    Fibonacci,
			DelayBetween: 10,
		},
		Cron:       "*/20 * * * * * *",
	}

	m, _ := json.Marshal(tt)
	fmt.Println(string(m))
	tt.Next5Times()
}

func TestLastSuccess(t *testing.T) {
	task := Task{}

	err := g.Save(&task).Error
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	time, err := findLastSuccessfulExecutionTime(task, g.Debug())
	spew.Dump(time, err)
}
