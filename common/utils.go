package common

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const localQueue string = "amqp://guest:guest@localhost:5672/"
const prodQueue string = "amqp://guest:guest@18.206.140.49:5672/"

// QueueConn is the current queue to connect to
const QueueConn string = prodQueue

// TimedJobType - json key for timed jobs
const TimedJobType = "timed"

// MonteCarloJobType - json key for monte-carlo jobs
const MonteCarloJobType = "monte-carlo"

// MergeSortJobType - json key for merge-sort jobs
const MergeSortJobType = "merge-sort"

// DevPort for server while in development
const devPort string = ":8090"
const prodPort string = ":80"

// Port used for accepting/sending requests to the server
const Port string = prodPort

const devEndpoint string = "localhost"
const prodEndpoint string = "localhost"

// Endpoint to listen/send HTTP requests
const Endpoint string = prodEndpoint

// Idle represents an idle job status
const Idle string = "idle"

// Working represents a working job status
const Working string = "working"

// Complete represents a completed job status
const Complete string = "completed"

// HeartRate is how often, in seconds, each node will pulse back to the server
const HeartRate time.Duration = 45

// Job represents an outgoing job to the consumers
type Job struct {
	ID   int
	Type string `json:"type"`
	Data []int  `json:"data"`
}

// Registration represents an imcoming request to register a worker
type Registration struct {
	Name string `json:"name"`
}

// WorkerStatus is the main "heartbeat" and confirmation between worker and server
type WorkerStatus struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
	LUD    time.Time `json:"lastUpdatedDate"`
	Name   string    `json:"name"`
}

// TimedJob represents a job to trigger the process to wait for "time"
type TimedJob struct {
	Time int `json:"time"`
}

// TrackJob represents a job that is currently being processed. This is used by the master to keep track of
// current running jobs
type TrackJob struct {
	ID        int       `json:"id"`
	Workers   int       `json:"numWorkers"`
	Results   []float64 `json:"results"`
	Completed int       `json:"completedJobs"`
	Status    string    `json:"jobStatus"`
	Type      string    `json:"jobType"`
}

// CompletedJob represents a response back from a node to the server
type CompletedJob struct {
	ID      int `json:"id"`
	Results []float64
}

// JobStatus represents a response from the server to the client regarding a job inquiry
type JobStatus struct {
	ID      int       `json:"id"`
	Results []float64 `json:"results"`
	Status  string    `json:"status"`
}

// InvalidJobError is called when json key "type" is not "timed" "monte-carlo" or "merge-sort"
func InvalidJobError() error {
	return errors.New("invalid job")
}
