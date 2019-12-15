package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/johnlahut/dist-go/common"
)

var id uuid.UUID
var name string
var status string
var seed rand.Source
var config common.Config

func getEndpoint() string {
	return "http://" + config.Host + config.Port
}

func monteCarlo(samples int) float64 {
	var m int
	rand.New(seed)
	for i := 0; i < samples; i++ {
		x := rand.Float64()
		y := rand.Float64()

		if x*x+y*y <= 1 {
			m++
		}
	}
	// time.Sleep(20 * time.Second)
	return (float64(m) / float64(samples)) * 4
}

func register() {

	// register worker
	body, err := json.Marshal(common.Registration{Name: name})
	common.FailOnError(err, "unable to register worker - request")

	resp, err := http.Post(getEndpoint()+"/register", "application/json", bytes.NewBuffer(body))
	common.FailOnError(err, "unable to register worker - response")

	// get response
	var confirm common.WorkerStatus
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&confirm)
	common.FailOnError(err, "unable to register worker - response")

	// set id
	id = confirm.ID
	log.Printf("successfully registered %s id: %s", name, id)
}

func pulse() {
	// send that we are still alive
	body, err := json.Marshal(common.WorkerStatus{Name: name, LUD: time.Now(), ID: id, Status: status})
	common.FailOnError(err, "unable to pulse server - request")

	// don't really care about the response
	_, err = http.Post(getEndpoint()+"/pulse", "application/json", bytes.NewBuffer(body))
	common.FailOnError(err, "unable to pulse server - response")
}

func complete(job common.Job, result []float64) {
	body, err := json.Marshal(common.CompletedJob{ID: job.ID, Results: result})
	common.FailOnError(err, "unable to complete job - request")

	// don't really care about the response
	_, err = http.Post(getEndpoint()+"/completed", "application/json", bytes.NewBuffer(body))
	common.FailOnError(err, "unable to complete job - request")

	log.Printf("computed result %f", result)
}

// Start will start consuming and processing from the queue
func main() {

	if len(os.Args) < 2 {
		common.FailOnError(errors.New(""), "usage: consumer [name]")
	}

	config = common.LoadConfig()

	name = os.Args[1]
	// connect to queue
	common.Connq()
	defer common.Closeq()

	q := common.Getq()
	ch := common.GetCh()

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	common.FailOnError(err, "Failed to register a consumer")

	register()

	heartbeat := time.NewTicker(common.HeartRate * time.Second)
	forever := make(chan bool)
	status = common.Idle
	seed = rand.NewSource(rand.Int63())

	// listen forever
	go func() {
		for d := range msgs {

			// unmarshal job
			var job common.Job
			err := json.Unmarshal(d.Body, &job)
			common.FailOnError(err, "failed to unmarshal job")
			status = common.Working
			pulse()
			// check job type
			switch job.Type {
			case common.TimedJobType:
				log.Printf("received timed job: %s", d.Body)

				t := time.Duration(job.Data[0])
				time.Sleep(t * time.Second)

				status = common.Idle
				pulse()
				complete(job, []float64{})
			case common.MonteCarloJobType:
				log.Printf("received monte-carlo job: %s", d.Body)
				result := monteCarlo(int(job.Data[0]))
				status = common.Idle

				pulse()
				complete(job, []float64{result})

			case common.MergeSortJobType:
				log.Printf("received merge-sort job: %s", d.Body)
				common.MergeSort(job.Data)
				status = common.Idle
				pulse()
				complete(job, job.Data)
			default:
				common.FailOnError(common.InvalidJobError(), "invalid job type")
			}

			// positive ack
			d.Ack(false)
			status = common.Idle
		}
	}()

	// update our server ever 15 seconds saying we are alive
	go func() {
		for {
			select {
			case <-heartbeat.C:
				pulse()
			}
		}
	}()

	log.Printf(" [*] waiting for messages. to exit press ctrl-c")
	<-forever
}
