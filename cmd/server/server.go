package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/johnlahut/dist-go/common"
)

var jobCount int
var workerCount int
var workers map[uuid.UUID]common.WorkerStatus
var activeJobs map[int]common.TrackJob
var queuedJobs map[int]common.Job

func sendToQueue(job common.Job) {

	msg, err := json.Marshal(&job)
	common.FailOnError(err, "failed to marshall message")
	common.Commitq(msg)

}

// function for registering workers
func register(w http.ResponseWriter, req *http.Request) {

	log.Printf("[$] %s", req.URL)

	// decode incoming registration request
	decoder := json.NewDecoder(req.Body)
	var reg common.Registration
	err := decoder.Decode(&reg)
	common.FailOnError(err, "failed to decode incoming registration request")

	// create unique id for job
	id := uuid.New()
	workers[id] = common.WorkerStatus{ID: id, Status: common.Idle, LUD: time.Now(), Name: reg.Name}

	// create response JSON
	confirm := common.WorkerStatus{ID: id, Status: common.Idle}
	res, err := json.Marshal(confirm)

	// send ack
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

	workerCount++
	log.Printf(">> number of workers: %d", workerCount)

	workers := activeWorkers()
	if workerCount == 1 && len(queuedJobs) != 0 && workers >= 1 {
		for _, v := range queuedJobs {
			delegateJob(v, workers)
		}
	}
}

// function for handling incoming job requests
func processJob(w http.ResponseWriter, req *http.Request) {

	log.Printf("[$] %s", req.URL)

	// decode incoming job request (from user)
	decoder := json.NewDecoder(req.Body)
	var job common.Job
	err := decoder.Decode(&job)
	common.FailOnError(err, "failed to decode incoming request")

	// add to job counter
	job.ID = jobCount
	jobCount++

	// pass off job to handler
	var res []byte
	workers := activeWorkers()
	if workers < 1 {
		log.Printf(">> no available workers. job %d is queued.", job.ID)
		queuedJobs[job.ID] = job
		confirm := common.JobStatus{ID: job.ID, Status: common.Queued}
		res, _ = json.Marshal(confirm)
	} else {
		delegateJob(job, workers)
		confirm := common.JobStatus{ID: job.ID, Status: common.Working}
		res, _ = json.Marshal(confirm)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func delegateJob(job common.Job, workers int) {

	log.Printf("sending %d job to queue", job.ID)

	activeJobs[job.ID] = common.TrackJob{
		ID: job.ID, Workers: workers, Results: []float64{}, Completed: 0,
		Status: common.Working, Type: common.MonteCarloJobType}

	switch job.Type {

	// splitting monte carlo job up
	case common.MonteCarloJobType:
		samples := job.Data[0] / workers
		carry := job.Data[0] % workers

		log.Printf(">> sending ~%d sample(s) to %d workers", samples, workers)

		for i := 0; i < workers; i++ {
			if i == workers-1 {
				samples += carry
			}
			sendToQueue(common.Job{ID: job.ID, Type: job.Type, Data: []int{samples}})
		}

	case common.MergeSortJobType:
		sendToQueue(job)
	case common.TimedJobType:
		sendToQueue(job)
	default:
		common.FailOnError(common.InvalidJobError(), "invalid job type")
	}
}

// return all registered workers
func serverInfo(w http.ResponseWriter, req *http.Request) {

	log.Printf("[$] %s", req.URL)

	var s []common.WorkerStatus
	for k, v := range workers {
		s = append(s, common.WorkerStatus{ID: k, Status: v.Status, LUD: v.LUD, Name: v.Name})
	}

	res, err := json.Marshal(s)
	common.FailOnError(err, "failed to decode incoming registration request")

	// send ack
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// each node is responsible for pulsing the server. if the server does not receive
// pulses for an extended period of time, the server will deregister the node
func jobPulse(w http.ResponseWriter, req *http.Request) {

	log.Printf("[$] %s", req.URL)

	decoder := json.NewDecoder(req.Body)
	var status common.WorkerStatus
	err := decoder.Decode(&status)
	common.FailOnError(err, "failed to decode pulse message from client")

	job, ok := workers[status.ID]
	if ok {
		job.LUD = time.Now()
		job.Status = status.Status
		workers[status.ID] = job
		log.Printf("[$] received pulse from %s status: [%s]", status.ID, status.Status)
	}
}

// each time a node completes a task, it will call this api endpoint
func completeJob(w http.ResponseWriter, req *http.Request) {
	log.Printf("[$] %s", req.URL)

	decoder := json.NewDecoder(req.Body)
	var completed common.CompletedJob
	err := decoder.Decode(&completed)
	common.FailOnError(err, "failed to decode completed job from client")

	job, ok := activeJobs[completed.ID]
	if ok {
		job.Results = append(job.Results, completed.Results[0])
		job.Completed++
	}

	// job is complete
	if job.Completed == job.Workers {
		switch job.Type {
		case common.MonteCarloJobType:
			var result float64
			for i := 0; i < job.Workers; i++ {
				result += job.Results[i]
			}
			job.Results = []float64{result / float64(job.Workers)}
			job.Status = common.Complete
		}
	}

	activeJobs[completed.ID] = job

}

func jobInfo(w http.ResponseWriter, req *http.Request) {
	log.Printf("[$] %s", req.URL)

	keys, ok := req.URL.Query()["id"]

	// respond with the status of a single job
	if ok {
		key, _ := strconv.Atoi(keys[0])
		job := activeJobs[key]
		res, err := json.Marshal(job)

		common.FailOnError(err, "failed to decode incoming registration request")
		// send ack
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)

		// respond with all workers processed in session
	} else {
		var s []common.TrackJob
		for _, v := range activeJobs {
			s = append(s, v)
		}

		res, err := json.Marshal(s)

		common.FailOnError(err, "failed to decode incoming registration request")
		// send ack
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}

}

func activeWorkers() (active int) {
	for _, v := range workers {
		if v.Status == common.Idle {
			active++
		}
	}
	return
}

func verify(tick *time.Ticker) {
	for {
		select {
		case <-tick.C:
			for k, v := range workers {
				if time.Now().Sub(v.LUD) > 45*time.Second {
					log.Printf("[!] detected node %s is unresponsive. removing from active nodes", v.ID)
					delete(workers, k)
					workerCount--
				}
			}
		}
	}
}

func main() {
	config := common.LoadConfig()
	log.Printf("[*] starting main node %s%s. to exit press ctrl-c", config.Host, config.Port)
	common.Connq()
	defer common.Closeq()

	workers = make(map[uuid.UUID]common.WorkerStatus)
	activeJobs = make(map[int]common.TrackJob)
	queuedJobs = make(map[int]common.Job)

	// endpoint mappings
	http.HandleFunc("/start", processJob)
	http.HandleFunc("/pulse", jobPulse)
	http.HandleFunc("/register", register)
	http.HandleFunc("/info", serverInfo)
	http.HandleFunc("/completed", completeJob)
	http.HandleFunc("/jobinfo", jobInfo)

	tick := time.NewTicker(common.VerifyRate * time.Second)
	go verify(tick)

	// start server
	http.ListenAndServe(config.Port, nil)
}
