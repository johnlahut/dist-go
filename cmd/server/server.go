package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/johnlahut/dist-go/common"
)

var jobCount int

func sendToQueue(job common.Job) {

	msg, err := json.Marshal(&job)
	common.FailOnError(err, "failed to marshall message")
	common.Commitq(msg)

}

// function for handling incoming job requests
func processJob(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var job common.Job
	err := decoder.Decode(&job)
	common.FailOnError(err, "failed to decode incoming request")

	job.ID = jobCount
	jobCount++

	log.Printf("sending %d job to queue", job.ID)

	switch job.Type {
	case common.TimedJobType, common.MonteCarloJobType, common.MergeSortJobType:
		sendToQueue(job)
	default:
		common.FailOnError(common.InvalidJobError(), "invalid job type")
	}
}

func main() {
	log.Println("[*] starting main node. to exit press ctrl-c")
	common.Connq()
	defer common.Closeq()

	// endpoint mappings
	http.HandleFunc("/start", processJob)

	// start server
	http.ListenAndServe(common.DevPort, nil)
}
