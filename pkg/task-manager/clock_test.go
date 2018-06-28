// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"testing"
	"time"
	"fmt"
)

func TestClock_AddListener(t *testing.T) {

	c := Clock{}

	i := func(ti time.Time) {
		fmt.Println(ti)
	}

	c.AddListener(&i, time.Duration(1*time.Second))
	c.Start()


	time.Sleep(10 * time.Second)

	c.Stop()
}
