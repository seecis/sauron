// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/jinzhu/gorm"
	"fmt"
	"encoding/json"
	"strconv"
	"time"
)

// tasks
//  create
//  read
//  update
//  delete
// orders
//  read
//  update
//  delete
// executions
//  read
//  update
//  delete

func CreateApi(g *gorm.DB) http.Handler {
	h := ManagerHandler{g: g}
	r := httprouter.New()
	r.GET("/task", h.TasksAll)
	r.PUT("/task", h.TaskCreate)
	r.POST("/task/:id", h.TaskUpdate)
	r.GET("/task/:id/detail", h.TaskDetail)
	r.GET("/task/:id", h.TaskSingle)
	return r
}

type ManagerHandler struct {
	g *gorm.DB
}

func (h *ManagerHandler) TaskCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer r.Body.Close()
	t := Task{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while parsing json", http.StatusBadRequest)
		return
	}

	err = h.g.Save(&t).Error
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while creating task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/task/%d", t.Id))
	w.WriteHeader(201)
}

func (h *ManagerHandler) TasksAll(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var tasks []Task
	err := h.g.Preload("Retry").Find(&tasks).Error
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while finding tasks", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *ManagerHandler) TaskSingle(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")

	t := Task{}
	u, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Bad id", http.StatusBadRequest)
		return
	}

	err = h.g.First(&t, u).Error
	if err != nil {
		b := gorm.IsRecordNotFoundError(err)
		if b {
			http.NotFound(w, r)
			return
		}

		fmt.Println(err)
		http.Error(w, "Error while fetching task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

type TaskDetail struct {
	LastExecution          *Execution `json:"last_execution"`
	LastSuccessful         *Execution `json:"last_successful"`
	LastFail               *Execution `json:"last_fail"`
	NextScheduledExecution *time.Time `json:"next_schedule"`
}

func findJoinedExecution(g *gorm.DB, taskID uint64, statuses ...ExecutionStatus) (*Execution, error) {
	mainJoin := joinIt(g, taskID)

	if len(statuses) > 0 {
		m := mainJoin.Where("executions.status = ?", int(statuses[0]))
		for i := 1; i < len(statuses); i++ {
			m = m.Or("executions.status = ?", int(statuses[i]))
		}

		mainJoin = m
	}

	e := Execution{}
	err := mainJoin.First(&e).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &e, nil
}

func joinIt(g *gorm.DB, taskID uint64) *gorm.DB {
	return g.Debug().
		Joins("JOIN execution_orders on execution_orders.id = executions.execution_order_id").
		Joins("JOIN tasks on tasks.id = execution_orders.task_id").
		Order("executions.time DESC").
		Where("tasks.id = ?", taskID)
}

func (h *ManagerHandler) TaskDetail(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	var last *Execution
	var lastSuccess *Execution
	var lastFail *Execution

	id := params.ByName("id")
	u, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Bad id", http.StatusBadRequest)
		return
	}

	last, err = findJoinedExecution(h.g, u, )
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while finding execution", http.StatusInternalServerError)
		return
	}

	lastSuccess, err = findJoinedExecution(h.g, u, Finished)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while finding execution", http.StatusInternalServerError)
		return
	}

	lastFail, err = findJoinedExecution(h.g, u, Failed)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while finding execution", http.StatusInternalServerError)
		return
	}

	t := Task{}
	err = h.g.First(&t, u).Error
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while finding execution details", http.StatusInternalServerError)
		return
	}

	na := t.NextAppointment()
	td := TaskDetail{
		LastExecution:          last,
		LastSuccessful:         lastSuccess,
		LastFail:               lastFail,
		NextScheduledExecution: &na,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(td)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error while getting execution", http.StatusInternalServerError)
		return
	}

	// Task itself
	// Retries
}

func (h *ManagerHandler) TaskUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.Error(w,"Not implemented", 501)
}
