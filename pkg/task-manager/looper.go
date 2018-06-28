// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"github.com/gorhill/cronexpr"
	"time"
	"fmt"
)

func(t *Task) FirstAppointment() time.Time {
	return cronexpr.MustParse(t.Cron).Next(time.Time{})
}

func (t *Task) NextAppointment() time.Time {
	return cronexpr.MustParse(t.Cron).Next(time.Now())
}

func (t *Task) Next5Times() bool {
	times := cronexpr.MustParse(t.Cron).NextN(time.Now(), 5)

	for e := range times {
		fmt.Printf("%v\n", times[e])
	}

	return false
}
