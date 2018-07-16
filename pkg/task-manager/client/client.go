// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package client

import (
	"github.com/seecis/sauron/pkg/task-manager"
	"net/http"
	"encoding/json"
	"bytes"
	"net/url"
	"fmt"
	"strconv"
)

type Client struct {
	httpClient *http.Client
	base       string
}

// Creates a new manager client. Use nill as httpclient to use default
func NewClient(baseUrl url.URL, httpClient *http.Client) (*Client) {
	u := baseUrl.String()
	h := http.DefaultClient
	if httpClient != nil {
		h = httpClient
	}

	return &Client{
		httpClient: h,
		base:       u,
	}
}


func isSuccessful(res *http.Response) bool {
	return (res.StatusCode >= 200) && (res.StatusCode < 400)
}

func (c *Client) CreateTask(t *task_manager.Task) (uint64, error) {
	u := c.base + "/task"
	b, err := json.Marshal(t)
	if err != nil {
		return 0, err
	}

	nr, err := http.NewRequest("PUT", u, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}

	nr.Header.Add("Content-Type", "application/json")
	res, err := c.httpClient.Do(nr)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	if isSuccessful(res) {
		s := res.Header.Get("Location")
		i := len("/task/")
		return strconv.ParseUint(s[i:], 10, 64)
	}

	return 0, fmt.Errorf("server returned %s", res.Status)
}

func (c *Client) GetTask(taskId uint64) (*task_manager.Task, error) {
	u := fmt.Sprintf("%s/task/%v", c.base, taskId)
	res, err := c.httpClient.Get(u)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if !isSuccessful(res) {
		return nil, fmt.Errorf("server returned %s", res.Status)
	}

	var t task_manager.Task
	err = json.NewDecoder(res.Body).Decode(&t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
