// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"net/http"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
	"os"
	"os/signal"
)

func Serve(ip, port string, db *gorm.DB) {
	c := Clock{}
	db.AutoMigrate(&Task{}, &ExecutionOrder{}, &Execution{}, &Retry{})

	defer db.Close()

	client := http.Client{}
	client.Timeout = 10 * time.Second
	ex := Executor{g: db, c: &client}
	lp := Looper{db, &ex}
	sh := Scheduler{g: db}

	looperCB := func(_ time.Time) { lp.onTick() }
	scheudlerCB := func(_ time.Time) { sh.Schedule() }
	c.AddListener(&looperCB, time.Second*2)
	c.AddListener(&scheudlerCB, time.Second*10)

	c.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)

	addr := fmt.Sprintf("%s:%s", ip, port)
	fmt.Println("Manager is listening on", port)
	h := &http.Server{Addr: addr, Handler: CreateApi(db)}


	go func() {
		err := h.ListenAndServe()
		fmt.Println(err)
	}()

	<-ch
	h.Close()
	fmt.Println("Received signal, closing")
	c.Stop()
}
