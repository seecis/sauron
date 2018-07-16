// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package client

import (
	"testing"
	"net/url"
	"github.com/seecis/sauron/pkg/task-manager"
	"fmt"
)

func TestClient_CreateTask(t *testing.T) {

	u, _ := url.Parse("http://localhost:8989")
	sc, err := NewClient(*u, nil)
	if err != nil {
		t.Error(err)
		return
	}

	newTaskUrl, err := sc.CreateTask(&task_manager.Task{
		Name:                 "Test Task",
		TriggerAddress:       "http://www.google.com",
		TriggerParams:        nil,
		TriggerMethod:        "GET",
		ErrorCallbackAddress: "",
		ErrorCallbackMethod:  "",
		Timeout:              0,
		Cron:                 "* * * * *",
		Disabled:             true,
	})

	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("A new task is created with id:", newTaskUrl)
}


func TestClient_GetTask(t *testing.T) {
	u, _ := url.Parse("http://localhost:8989")
	sc, err := NewClient(*u, nil)
	if err != nil {
		t.Error(err)
		return
	}

	task, err := sc.GetTask(1)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(task)
}
