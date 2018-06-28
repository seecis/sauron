// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/gorhill/cronexpr"
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"math"
)

type RetryType int

const (
	None        RetryType = iota
	Fibonacci
	Exponential
	Plain
)

func (r RetryType) MarshalJSON() ([]byte, error) {
	var exp string
	switch r {
	case Fibonacci:
		exp = "fibonacci"
	case Exponential:
		exp = "exponential"
	case Plain:
		exp = "plain"
	case None:
		exp = "none"
	}

	return json.Marshal(exp)
}

func (r *Retry) UnmarshalJSON(b []byte) error {
	type rRetry struct {
		Id           uint64        `json:"id"`
		RetryType    string        `json:"retry_type"`
		DelayBetween time.Duration `json:"delay_between"`
		MaxRetries   int           `json:"max_retries"`
	}

	rr := rRetry{}

	err := json.Unmarshal(b, &rr)
	if err != nil {
		return err
	}

	r.Id = rr.Id
	r.RetryType = parseRetryType(rr.RetryType)
	r.DelayBetween = rr.DelayBetween
	r.MaxRetries = rr.MaxRetries

	return nil
}
func parseRetryType(s string) RetryType {
	switch s {
	case "fibonacci":
		return Fibonacci
	case "exponential":
		return Exponential
	case "plain":
		return Plain
	case "none":
		return None
	}

	return None
}

type Retry struct {
	Id           uint64        `json:"id"`
	RetryType    RetryType     `json:"retry_type"`
	DelayBetween time.Duration `json:"delay_between"`
	MaxRetries   int           `json:"max_retries"`
}

type mapp map[string]interface{}

func (a mapp) Value() (driver.Value, error) {
	v, e := json.Marshal(a)
	if e != nil {
		return nil, e
	}

	return string(v), nil
}

func (ap *mapp) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("Incompatible type for map")
	}

	return json.Unmarshal(source, ap)
}

type Task struct {
	Id                   uint64           `json:"id"`
	Name                 string           `json:"name"`
	TriggerAddress       string           `json:"trigger_address"`
	TriggerParams        mapp             `gorm:"type:varchar(1250)" json:"trigger_params"`
	TriggerMethod        string           `json:"trigger_method"`
	ErrorCallbackAddress string           `json:"error_callback_address"`
	ErrorCallbackMethod  string           `json:"error_callback_method"`
	Timeout              uint             `json:"timeout"`
	Retry                Retry            `gorm:"foreignKey:RetryId" json:"retry"`
	RetryId              uint64           `json:"-"`
	Cron                 string           `json:"cron"`
	ExecutionOrders      []ExecutionOrder `gorm:"foreignKey:TaskId" json:"-"`
	Disabled             bool             `json:"disabled"`
}

func (task Task) NextAppointmentAfter(i time.Time) time.Time {
	return cronexpr.MustParse(task.Cron).Next(i)
}

type ExecutionStatus int

const (
	Pending   ExecutionStatus = iota
	Finished
	Failed
	Running
	Cancelled
)

func (es ExecutionStatus) String() string {
	switch es {
	case Finished:
		return "Finished"
	case Failed:
		return "Failed"
	case Running:
		return "Running"
	case Cancelled:
		return "Cancelled"
	case Pending:
		return "Pending"
	}

	return "unknown"
}

type ExecutionOrder struct {
	Id         uint64
	Status     ExecutionOrderStatus
	TaskId     uint64
	Executions []Execution `gorm:"foreignKey:ExecutionOrderId"`
	Expire     time.Time
}

func (order ExecutionOrder) findLastExecution() *Execution {
	t := time.Time{}

	if order.Executions == nil {
		return nil
	}

	key := 0
	executions := order.Executions
	for k, v := range executions {
		if v.Date.After(t) {
			t = v.Date
			key = k
		}
	}

	return &executions[key]
}

func (order ExecutionOrder) Ready(retry Retry) (bool) {
	le := order.findLastExecution()
	if le == nil {
		return true
	}

	rc := len(order.Executions)

	var nextRetryTime time.Time
	switch retry.RetryType {
	case Fibonacci:
		fib := FibonacciIndex(rc)
		nextRetryTime = le.Date.Add(time.Duration(fib * int(retry.DelayBetween.Seconds())))
	case Exponential:
		b := math.Pow(2, float64(rc)) - 1
		nextRetryTime = le.Date.Add(time.Duration(b * retry.DelayBetween.Seconds()))
	case Plain:
		nextRetryTime = le.Date.Add(time.Duration(float64(rc) * retry.DelayBetween.Seconds()))
	case None:
		fmt.Println("We shouldn't be here")
		panic("Trying to retry an order which mustn't be retried")
		return false
	}

	fmt.Println("Task will be run on", nextRetryTime)
	return nextRetryTime.Before(time.Now())
}

type Execution struct {
	Id               uint64          `json:"id"`
	Status           ExecutionStatus `json:"status"`
	Date             time.Time       `json:"date"`
	Message          string          `json:"message"`
	ExecutionOrderId uint64          `json:"execution_order_id"`
	// milliseconds
	Took uint32 `json:"took"`
}

type ExecutionOrderStatus int

const (
	Scheduled    ExecutionOrderStatus = iota
	Dispatched
	Fulfilled
	NotFulfilled
	WontFulfill
	Expired
)

type Scheduler struct {
	g *gorm.DB
}

func findLastSuccessfulExecutionTime(task Task, g *gorm.DB) (time.Time, error) {
	t := Task{}
	err := g.Preload("ExecutionOrders", "status = ?", Fulfilled).
		Preload("ExecutionOrders.Executions", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", Finished).Order("executions.time desc").Limit(1)
	}).First(&t, "id = ?", task.Id).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return time.Time{}, nil
		}

		return time.Time{}, err
	}

	return t.ExecutionOrders[0].Executions[0].Date, nil
}

func (s *Scheduler) Schedule() error {
	t := time.Now()
	fmt.Println("Scheduler tick")

	rowCount := s.g.Model(&ExecutionOrder{}).Where("expire < ? AND status != ?", t, Expired).Update("status", Expired).RowsAffected
	e := s.g.Error
	if e != nil {
		return e
	}

	fmt.Printf("Set %d orders expired\n", rowCount)
	newTasks, err := s.findReadyTasks()
	if err != nil {
		return err
	}

	// Create an execution order for each of them.
	for _, v := range *newTasks {
		if err != nil {
			fmt.Println(err)
		}

		e := ExecutionOrder{
			Status: Scheduled,
			TaskId: v.Id,
			Expire: v.NextAppointmentAfter(t),
		}

		if err := s.g.Save(&e).Error; err != nil {
			return err
		}

		fmt.Printf("Task with id %d scheduled with order %d \n", e.TaskId, e.Id)
	}

	return nil
}

func (s ExecutionOrderStatus) IsFinished() bool {
	return !(s == Scheduled || s == Dispatched || s == NotFulfilled)
}

func isAllFinished(eos []ExecutionOrder) bool {
	for _, o := range eos {
		s := o.Status
		if !s.IsFinished() {
			return false
		}
	}

	return true
}

func (s *Scheduler) findReadyTasks() (*[]Task, error) {
	var tasks []Task
	err := s.g.Find(&tasks, "disabled = ?", false).Error
	if err != nil {
		return nil, err
	}

	var ut []Task
	for _, t := range tasks {
		var eos []ExecutionOrder
		err := s.g.Model(&t).Related(&eos).Error
		if err != nil {
			return nil, err
		}

		// todo: extract this to sql?
		if !isAllFinished(eos) {
			continue
		}

		ut = append(ut, t)
	}

	return &ut, err
}

type Looper struct {
	g *gorm.DB
	e *Executor
}

type ExecutionResult struct {
	err         error
	taskId      uint64
	orderId     uint64
	executionId uint64
	status      ExecutionStatus
}

func filterOrderById(orders *[]ExecutionOrder, id uint64) *ExecutionOrder {
	for _, v := range *orders {
		if v.Id == id {
			return &v
		}
	}

	return nil
}

func (order ExecutionOrder) GetTask(g *gorm.DB) (*Task, error) {
	t := Task{}
	err := g.First(&t, "id = ?", order.TaskId).Error
	if err != nil {
		// wat do?
		return nil, err
	}

	return &t, nil
}
func (order ExecutionOrder) retryCount(g *gorm.DB) (int, error) {
	c := 0
	if err := g.Count(&c).Model(&Execution{}).Where("execution_order_id = ?", order.Id).Error; err != nil {
		return -1, err
	}

	return c, nil
}

func (order ExecutionOrder) SetUnschedulable(db *gorm.DB) error {
	order.Status = WontFulfill
	return db.Save(&order).Error
}

func (l *Looper) onTick() error {
	var orders []ExecutionOrder
	if err := l.g.Find(&orders, "status = ? OR status = ? OR status = ?", Pending, NotFulfilled, Cancelled).Error; err != nil {
		return err
	}

	if len(orders) == 0 {
		fmt.Println("Nothing to do")
		return nil
	}

	er := make(chan ExecutionResult)
	pendingOrderIds := make(map[uint64]bool)
	for _, v := range orders {
		t, err := v.GetTask(l.g)
		if err != nil {
			return err
		}

		if !v.Ready(t.Retry) {
			continue
		}

		rc := len(v.Executions)
		if rc > t.Retry.MaxRetries {
			err := v.SetUnschedulable(l.g)
			if err != nil {
				fmt.Println(err)
			}
		}

		// wait for silent period

		pendingOrderIds[v.Id] = true
		l.e.ExecuteAsync(v, er)
		l.g.Save(v)
	}

	for len(pendingOrderIds) > 0 {
		e := <-er
		delete(pendingOrderIds, e.orderId)
		o := filterOrderById(&orders, e.orderId)
		if o == nil {
			fmt.Println("Coudn't find order??")
			continue
		}

		fmt.Printf("Task %d, Order %d, Execution %d", o.TaskId, o.Id, e.executionId)
		switch e.status {
		case Pending:
			o.Status = Dispatched
			fmt.Printf(" is dispatched\n")
		case Running:
			o.Status = Dispatched
			fmt.Printf(" is dispatched\n")
		case Finished:
			o.Status = Fulfilled
			fmt.Printf(" is fulfilled\n")
		case Failed:
			o.Status = NotFulfilled
			fmt.Printf(" is not fulfilled reason: task is failed\n")
		case Cancelled:
			o.Status = NotFulfilled
			fmt.Printf(" is not fulfilled reason: execution is cancelled\n")
		}

		err := l.g.Save(o).Error
		if err != nil {
			fmt.Println(err)
			// todo: handle this too
		}
	}

	return nil
}

type Executor struct {
	g *gorm.DB
	c *http.Client
}

func (executor *Executor) ExecuteAsync(order ExecutionOrder, c chan ExecutionResult) () {
	go func() {
		e := Execution{
			Status:           Pending,
			Date:             time.Now(),
			Message:          "Starting",
			ExecutionOrderId: order.Id,
		}

		result := ExecutionResult{
			err:         nil,
			taskId:      order.TaskId,
			orderId:     order.Id,
			executionId: e.Id,
			status:      Failed,
		}

		err := executor.g.Save(&e).Error
		if err != nil {
			result.err = err
			c <- result
			return
		}

		result.executionId = e.Id
		err = executor.execute(order, &e)
		if err != nil {
			result.err = err
			c <- result
			return
		}

		result.status = Finished
		result.err = nil
		c <- result
	}()
}

func (executor *Executor) executionFailed(ex *Execution, msg string) {
	fmt.Println("Execution with id ", ex.Id, "failed: ", msg)
	ex.Status = Failed
	ex.Message = msg
	executor.g.Save(ex)
}

func (executor *Executor) execute(order ExecutionOrder, ex *Execution) error {
	var err error
	t, err := order.GetTask(executor.g)
	client := http.Client{
		Timeout: time.Second * time.Duration(t.Timeout),
	}

	if err != nil {
		return err
	}

	startTime := time.Now()

	defer func(e *Execution) {
		if r := recover(); r != nil {
			executor.executionFailed(e, fmt.Sprintf("%v", r))
		}
	}(ex)

	defer func(e *Execution) {
		if err != nil {
			executor.executionFailed(e, (err).Error())
		}
	}(ex)

	r, err := http.NewRequest(t.TriggerMethod, t.TriggerAddress, nil)
	if err != nil {
		return err
	}

	for k, v := range t.TriggerParams {
		var vString string
		switch v.(type) {
		case int:
			vString = string(v.(int))
		case string:
			vString = v.(string)
		}

		r.URL.Query().Add(k, vString)
	}

	res, err := client.Do(r)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 && res.StatusCode >= 400 {
		msg := res.Header.Get("Message")
		ex.Message = msg
		ex.Status = Failed
	} else {
		ex.Status = Finished
		ex.Message = "Success"
	}

	finishTime := time.Now()
	took := finishTime.Sub(startTime)
	ex.Took = uint32(took / time.Millisecond)
	return executor.g.Save(ex).Error
}
