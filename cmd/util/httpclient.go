// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"github.com/seecis/sauron/pkg/scheduler"
	"bytes"
	"fmt"
	"time"
	"encoding/json"
	"strings"
	"github.com/seecis/sauron/pkg/extractor"
	http2 "github.com/seecis/sauron/internal/http"
)

var DefaultSauronHttpClient = SauronHttpClient{
	urlBase: "http://localhost:9091/",
	client: http.Client{
		Timeout: 10 * time.Second,
	},
}

type SauronHttpClient struct {
	urlBase string
	client  http.Client
}

func (s *SauronHttpClient) ScheduleExtraction(extractorId string, url string) (string, error) {
	var endpoint = s.urlBase + "extract/" + extractorId
	body := scheduler.ExtractionRequest{
		Url: url,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return "", err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("Error while scheduling. Error code: %d\n", res.StatusCode)
	}

	locHeader := res.Header.Get("Location")
	i := strings.Index(locHeader, "/report/")
	return locHeader[i+len("/report/"):], nil
}

func (s *SauronHttpClient) AddExtractor(extractor extractor.Extractor) (string, error) {
	var endpoint = s.urlBase + "extractor/"
	b, err := json.Marshal(extractor)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(b))
	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}

	locHeader := res.Header.Get("Location")
	heading := "/extractor/"
	i := strings.Index(locHeader, heading)
	return locHeader[i+len(heading):], nil
}
func (s *SauronHttpClient) GetExtractors() ([]extractor.Extractor, error) {
	var endpoint = s.urlBase + "extractor/"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var ex []extractor.HtmlExtractor
	err = json.NewDecoder(res.Body).Decode(&ex)
	if err != nil {
		return nil, err
	}

	var q []extractor.Extractor
	for _, v := range ex {
		q = append(q, v)
	}

	return q, nil
}
func (s *SauronHttpClient) GetReportHeaders() ([]http2.ApiReport, error) {
	var endpoint = s.urlBase + "report/"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	var apiReports []http2.ApiReport
	err = json.NewDecoder(res.Body).Decode(&apiReports)
	return apiReports, err
}
