// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"time"
)

type clockListener struct {
	listener     *func(time.Time)
	interval     time.Duration
	lastTickSent time.Time
}

type Clock struct {
	listeners []clockListener
	ticker    *time.Ticker
	stop      bool
}

func (c *Clock) Start() {
	if c.ticker != nil {
		return
	}
	ticker := time.NewTicker(time.Second)

	go func() {
		for !c.stop {
			tick := <-ticker.C
			for k, v := range c.listeners {
				theTime := v.lastTickSent.Add(v.interval)
				if theTime.Before(time.Now()) {
					vptr := &v
					vptr.lastTickSent = tick
					c.listeners[k].lastTickSent = tick
					l := *v.listener
					l(tick)
				}
			}
		}
	}()
}

func (c *Clock) AddListener(listener *func(time.Time), interval time.Duration) {
	c.listeners = append(c.listeners, clockListener{
		listener:     listener,
		interval:     interval,
		lastTickSent: time.Time{},
	})
}
func (c *Clock) Stop() {
	c.stop = true
}
