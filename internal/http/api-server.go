// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

///extractor
//	create
//		Accepts: json, yaml
//	read
//		Produces: json, yaml
///extractor/{id}
//	read
//		Produces: json, yaml
//	delete
//		Just delete
//	update
//		Accepts: json, yaml
//		Does not return the extractor
///extract?extractorId={"extractorId"}
//	POST
//	Accepts: json, yaml
//	Example:
//	{
//		"url":["example.com"]
//	}
//	Returns: [json, yaml], Created report Uid, status
///report/{reportId}
//	create ??
//	read
//	update ??
//	delete

package http

import (
	"encoding/json"
	"fmt"
	"github.com/seecis/sauron/internal/dataaccess"
	"github.com/seecis/sauron/pkg/extractor"
	"github.com/seecis/sauron/pkg/scheduler"
	"github.com/julienschmidt/httprouter"
	"github.com/gorilla/handlers"
	"io"
	"os"
	"log"
	"mime"
	"net/http"
	"strconv"
	"github.com/rs/cors"
	"github.com/RichardKnop/machinery/v1"
	machinery2 "github.com/seecis/sauron/internal/machinery"
	"github.com/segmentio/ksuid"
	"time"
	"github.com/spf13/viper"
	"github.com/davecgh/go-spew/spew"
	"github.com/seecis/sauron/pkg/task-manager/client"
	"github.com/seecis/sauron/pkg/task-manager"
	"net/url"
	"github.com/jinzhu/gorm"
)

func ServeApi(ip, port string) {

	spew.Dump(viper.AllKeys())
	// todo add this to config.
	//fses := dataaccess.NewFileSystemExtractorService("extractors/")
	mses := dataaccess.NewMsSqlExtractorService(true, false)
	rserv := dataaccess.NewMSSQLReportService(true, true)
	jserv := dataaccess.NewMSSQLJobService(true, true)

	u, _ := url.Parse(viper.GetString("serve.scheduler_addr"))
	eh := &ExtractorHandler{service: mses,
		scheduler: &HtmlExtractorScheduler{machinery: machinery2.NewMachinery()},
		reportService: rserv,
		jobsService: jserv,
		schedulerClient: client.NewClient(*u, nil)}

	router := httprouter.New()
	router.GET("/", index)
	router.GET("/extractor", eh.GetAll)
	router.PUT("/extractor", eh.NewExtractor)

	router.GET("/extractor/:id", eh.GetExtractor)
	router.DELETE("/extractor/:id", eh.DeleteExtractor)
	router.POST("/extractor/:id", eh.UpdateExtractor)
	router.POST("/extract/:id", eh.Extract)
	router.GET("/report/:id", eh.GetReport)
	router.HEAD("/report/", eh.GetReportHeaders)
	router.GET("/report/", eh.GetReports)
	router.GET("/job/", eh.GetAllJobs)
	router.GET("/job/:id", eh.GetJob)
	router.POST("/job/", eh.CreateJob)
	router.POST("/start/job/:id", eh.StartJob)

	address := fmt.Sprintf("%s:%s", ip, port)

	accessLogFile := createAccessLog()
	defer accessLogFile.Close()
	multiout := io.MultiWriter(accessLogFile, os.Stdout)

	mux := http.NewServeMux()
	mux.Handle("/", router)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowCredentials: true,
		Debug:            false,
		ExposedHeaders:   []string{"Location"}})

	handler := c.Handler(mux)
	log.Println("Sauron api is listening at ", address)
	log.Fatal(http.ListenAndServe(address, handlers.LoggingHandler(multiout, handler)))
}

func checkFileExists(path string) bool {
	//os.Stat check for file info. If Stat couldn't find a file returns an error. Since golang doesn't provide an error
	//handling mechanism os library provides us with a nice IsNotExists function. IsNotExists(error) checks if returned
	//error occurred because of an file not found error. Since we don't care about the file info we just ignore Stat's
	//first return value.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func createAccessLog() *os.File {
	// Todo: add to config
	basePath := "access_log"
	extension := ".txt"

	created := false

	path := basePath + extension
	for i := 1; created == false; i++ {
		exists := checkFileExists(path)
		if exists {
			path = basePath + strconv.Itoa(i) + extension
			created = false
			continue
		}

		created = true
	}

	f, err := os.Create(path)
	if err != nil {
		// Todo: What to do here?
		log.Panic(err)
	}

	return f
}

func index(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprintln(w, "Sauron is watching you")
	fmt.Fprintln(w, "You requested "+r.URL.Path)
	fmt.Fprintln(w, params)
}

type mimeType int

const (
	mime_json = iota
)

func getMimeType(typeHeader string) mimeType {
	mediaType, _, err := mime.ParseMediaType(typeHeader)
	if err != nil {
		return mime_json
	}

	switch mediaType {
	case "text/json":
		return mime_json
	}

	return mime_json
}

func getResponseType(r *http.Request) mimeType {
	// todo: test this EXHAUSTIVELY
	s := r.Header.Get("Accept")
	return getMimeType(s)
}

func getContentType(r *http.Request) mimeType {
	s := r.Header.Get("Content-Type")
	return getMimeType(s)
}

type ExtractorHandler struct {
	service         dataaccess.ExtractorService
	reportService   dataaccess.ReportService
	scheduler       scheduler.ExtractionScheduler
	jobsService     dataaccess.JobService
	schedulerClient *client.Client
}

func (eh *ExtractorHandler) GetAll(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	allExtractors, err := eh.service.GetAll()
	if err != nil {
		log.Panic(err)
	}

	w.Header().Set("Content-Type", "text/json")
	json.NewEncoder(w).Encode(allExtractors)
}

type Unmarshaller interface {
	Decode(interface{}) error
}

func serialize(w http.ResponseWriter, thing interface{}, mimeType mimeType) error {
	return json.NewEncoder(w).Encode(thing)
}

func deserialize(r io.Reader, htmlExtractor *extractor.HtmlExtractor) error {
	decode := json.NewDecoder(r).Decode(htmlExtractor)
	return decode
}

func (eh *ExtractorHandler) NewExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Todo: we may need more magic here
	var ex extractor.HtmlExtractor
	err := deserialize(r.Body, &ex)
	defer r.Body.Close()
	if err != nil {
		// Todo: make this warning
		log.Println("Error while unmarshalling :" + err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id, se := eh.service.Save(&ex)
	if se != nil {
		log.Printf(se.Error())
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/extractor/"+id)
	w.WriteHeader(http.StatusCreated)
}

func (eh *ExtractorHandler) GetExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("id")
	ex, srvErr := eh.service.Get(name)
	if srvErr != nil {
		if dataaccess.IsNotFound(srvErr) {
			http.NotFound(w, r)
			return
		}

		log.Println("Cannot get extractor with id: " + name)
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	serialize(w, ex, getResponseType(r))
}

func (eh *ExtractorHandler) DeleteExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	extractorId := params.ByName("id")
	if extractorId == "" {
		http.Error(w, "Bad request: id is needed on request path", http.StatusBadRequest)
		return
	}

	se := eh.service.Delete(extractorId)
	if se != nil {
		if dataaccess.IsNotFound(se) {
			http.NotFound(w, r)
		}
	}
}

func (eh *ExtractorHandler) UpdateExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("id")
	if name == "" {
		http.Error(w, "Bad request: id is needed on request path", http.StatusBadRequest)
		return
	}

	ex, err := eh.service.Get(name)
	if err != nil {
		if dataaccess.IsNotFound(err) {
			http.NotFound(w, r)
			return
		}

		log.Println(err.Error())
		if dataaccess.IsBadRequest(err) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	var newEx extractor.HtmlExtractor
	err = deserialize(r.Body, &newEx)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error while deserializing request", http.StatusBadRequest)
		return
	}

	err = eh.service.Delete(ex.GetName())
	if err != nil {
		log.Panic(err)
		return
	}

	newEx.Name = ex.GetName()
	_, err = eh.service.Save(&newEx)
	if err != nil {
		log.Panic(err)
	}
}

func (eh *ExtractorHandler) Extract(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	extractorId := params.ByName("id")
	if extractorId == "" {
		http.Error(w, "Needs to provide an extractor id", http.StatusBadRequest)
		return
	}

	ex, err := eh.service.Get(extractorId)
	if dataaccess.IsNotFound(err) {
		http.NotFound(w, r)
		return
	}

	um := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var e scheduler.ExtractionRequest
	err = um.Decode(&e)
	if err != nil {
		http.Error(w, "Malformed payload", http.StatusBadRequest)
		return
	}

	reportId, err := eh.scheduler.Schedule(ex.GetId(), e)
	if err != nil {
		http.Error(w, "Error while scheduling extractor for request", http.StatusInternalServerError)
		log.Println("Error while scheduling", err, e)
		return
	}

	w.Header().Set("Location", "/report/"+reportId)
	w.WriteHeader(http.StatusAccepted)
	return
}

type HtmlExtractorScheduler struct {
	machinery *machinery.Server
}

func (hes *HtmlExtractorScheduler) ScheduleSync(extractorId uint64, payload scheduler.ExtractionRequest) (string, error) {
	ej := machinery2.NewExtractionJob(payload.Url, extractorId, payload.ReportId)
	a, err := hes.machinery.SendTask(ej)

	if err != nil {
		log.Fatal(err)
	}

	a.GetWithTimeout(time.Second*1, time.Minute*1)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (hes *HtmlExtractorScheduler) Schedule(extractorId uint64, payload scheduler.ExtractionRequest) (string, error) {
	k := ksuid.New()
	ej := machinery2.NewExtractionJob(payload.Url, extractorId, payload.ReportId)
	_, err := hes.machinery.SendTask(ej)
	if err != nil {
		log.Fatal(err)
	}

	return k.String(), nil
}

func (eh *ExtractorHandler) GetReports(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	allReports, err := eh.reportService.GetAll()
	if err != nil {
		http.Error(w, "Error while handling request", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allReports)
}

func (eh *ExtractorHandler) GetReport(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	_, err := ksuid.Parse(id)
	if err != nil {
		http.Error(w, "Error while parsing id", http.StatusBadRequest)
		return
	}

	report, err := eh.reportService.Get(id)
	k, err := ksuid.FromBytes(report.UID)
	res := struct {
		Id    string `json:"id"`
		Field dataaccess.Field
	}{
		k.String(),
		report.Field,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, "Error while handling request", http.StatusInternalServerError)
		return
	}
}

func (eh *ExtractorHandler) GetJob(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")

	k, err := ksuid.Parse(id)
	if err != nil {
		http.Error(w, "Error while parsing id", http.StatusBadRequest)
		log.Println(err)
		return
	}

	j, err := eh.jobsService.Get(k.String())
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "Error while parsing id", http.StatusBadRequest)
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")

	kk, err := ksuid.FromBytes(j.UID)
	j.Ksuid = kk.String()
	json.NewEncoder(w).Encode(j)
}

func (eh *ExtractorHandler) GetAllJobs(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	jobs, err := eh.jobsService.GetAll()

	w.Header().Set("Content-Type", "application/json")

	for e := range jobs {
		j := jobs[e]
		kk, err := ksuid.FromBytes(j.UID)
		if err != nil {
			http.Error(w, "Error while handling request", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		j.Ksuid = kk.String()
	}

	err = json.NewEncoder(w).Encode(jobs)

	if err != nil {
		http.Error(w, "Error while handling request", http.StatusInternalServerError)
		//Todo: log this
		return
	}
}

type NewJob struct {
	ExtractorId string `json:"extractorId"`
	Urls        []string
	Cron        string
}

func (eh *ExtractorHandler) StartJob(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	jobId := params.ByName("id")

	job, err := eh.jobsService.Get(jobId)

	if err != nil {
		http.Error(w, "Error while getting the job", 404)
		log.Println(err)
		return
	}

	urls := job.Urls

	for k := range urls {
		r, err := eh.reportService.CreateForJob(job, nil)
		if err != nil {
			log.Println(err)
			continue
		}

		eh.scheduler.ScheduleSync(job.HtmlExtractor.ID,
			scheduler.ExtractionRequest{
				Url:      urls[k].Url,
				ReportId: r.ID,
			})

		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func (eh *ExtractorHandler) CreateJob(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer r.Body.Close()
	var nj NewJob
	err := json.NewDecoder(r.Body).Decode(&nj)
	if err != nil {
		http.Error(w, "Malformed input", http.StatusBadRequest)
		log.Println(err)
		return
	}

	e, err := eh.service.Get(nj.ExtractorId)
	if err != nil {
		http.Error(w, "Extractor not found", http.StatusNotFound)
		log.Println(err)
		return
	}

	var urls []*dataaccess.Url

	for e := range nj.Urls {
		urls = append(urls, &dataaccess.Url{
			Url: nj.Urls[e],
		})
	}

	j := &dataaccess.Job{
		Urls:            urls,
		HtmlExtractorId: (e.(extractor.HtmlExtractor)).Id,
		Cron:            nj.Cron,
	}

	jobId, err := eh.jobsService.Save(j)

	triggerAddress := viper.GetString("serve.self_url") + "start/job/" + jobId
	t := task_manager.Task{
		TriggerAddress:       triggerAddress,
		TriggerParams:        nil,
		TriggerMethod:        "",
		ErrorCallbackAddress: "",
		ErrorCallbackMethod:  "",
		Timeout:              60,
		Retry: task_manager.Retry{
			RetryType:    task_manager.Fibonacci,
			DelayBetween: 10,
			MaxRetries:   2,
		},
		Cron:     nj.Cron,
		Disabled: false,
	}

	scheduledId, err := eh.schedulerClient.CreateTask(&t)
	if err != nil {
		http.Error(w, "Error while scheduling", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	j.ScheduledTaskId = scheduledId
	_, err = eh.jobsService.Save(j)
	if err != nil {
		http.Error(w, "Error while rescheduling", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Location", "/job/"+jobId)
	w.WriteHeader(http.StatusCreated)
}

func (eh *ExtractorHandler) GetReportHeaders(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	reports, err := eh.reportService.GetHeaders()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error while getting reports", http.StatusInternalServerError)
		return
	}

	var apiReports []ApiReport

	for _, v := range reports {
		k, err := ksuid.FromBytes(v.UID)
		if err != nil {
			http.Error(w, "Error while handling db results", http.StatusInternalServerError)
			return
		}

		apiReports = append(apiReports, ApiReport{
			UID:       k,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			DeletedAt: v.DeletedAt,
		})
	}

	if apiReports == nil {
		apiReports = []ApiReport{}
	}

	// Todo: make this work with yaml also
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(apiReports)
	if err != nil {
		http.Error(w, "An error occured while marshaling", http.StatusInternalServerError)
		return
	}
}

type ApiReport struct {
	UID       ksuid.KSUID `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	DeletedAt *time.Time  `json:"deleted_at"`
}
