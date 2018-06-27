// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"github.com/jinzhu/gorm"
	"time"
	"os"
	"os/signal"
	"fmt"
	"net/http"
)

func Main() {
	c := Clock{}

	g, err := gorm.Open("mssql", "sqlserver://sa:AAABBBccc123@192.168.99.100:1433?database=white_wizard")
	if err != nil {
		panic(err)
	}

	g.AutoMigrate(&Task{}, &ExecutionOrder{}, &Execution{}, &Retry{})

	defer g.Close()

	client := http.Client{}
	client.Timeout = 10 * time.Second
	ex := Executor{g: g, c: &client}
	lp := Looper{g, &ex}
	sh := Scheduler{g: g}

	looperCB := func(_ time.Time) { lp.onTick() }
	scheudlerCB := func(_ time.Time) { sh.Schedule() }
	c.AddListener(&looperCB, time.Second*2)
	c.AddListener(&scheudlerCB, time.Second*10)

	c.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)

	h := &http.Server{Addr: "0.0.0.0:9868", Handler: CreateApi(g)}


	go func() {
		err := h.ListenAndServe()
		fmt.Println(err)
	}()

	<-ch
	h.Close()
	fmt.Println("Received signal, closing")
	c.Stop()
}
