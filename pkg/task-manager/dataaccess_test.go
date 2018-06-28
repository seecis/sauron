// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"testing"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"

	"log"
	"fmt"
	"net/http"
	"time"
)

var g *gorm.DB
var s *Scheduler

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tearDown()
}

func tearDown() {
	g.Close()
}

func setup() {
	var err error
	g, err = gorm.Open("mssql", "sqlserver://sa:AAABBBccc123@192.168.99.100:1433?database=white_wizard")
	if err != nil {
		log.Fatal(err)
	}
}

func TestScheduler_OnTick(t *testing.T) {
	g.AutoMigrate(&Task{}, &ExecutionOrder{}, &Execution{}, &Retry{})
	s = &Scheduler{g: g}

	t1 := Task{
		Name:                 "Test1",
		TriggerAddress:       "test.example.com",
		TriggerMethod:        "GET",
		ErrorCallbackAddress: "error.example.com",
		TriggerParams: map[string]interface{}{
			"aa": "bb",
		},
		ErrorCallbackMethod: "POST",
		Timeout:             30,
		Retry: Retry{
			RetryType:    Plain,
			DelayBetween: 30,
		},
		Cron:       "*/2 * * * * * *",
	}

	err := g.Save(&t1).Error
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	s.Schedule()
}

func TestLooper_OnTick(t *testing.T) {

	cc := http.Client{
		Timeout: 50 * time.Second,
	}

	l := Looper{
		g: g,
		e: &Executor{
			g: g,
			c: &cc,
		},
	}
	err := l.onTick()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}
