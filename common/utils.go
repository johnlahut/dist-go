package common

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
)

// TimedJobType - json key for timed jobs
const TimedJobType = "timed"

// MonteCarloJobType - json key for monte-carlo jobs
const MonteCarloJobType = "monte-carlo"

// MergeSortJobType - json key for merge-sort jobs
const MergeSortJobType = "merge-sort"

// Idle represents an idle job status
const Idle string = "idle"

// Working represents a working job status
const Working string = "working"

// Complete represents a completed job status
const Complete string = "completed"

// Queued represents a job that has no available workers
const Queued string = "queued"

// HeartRate is how often, in seconds, each node will pulse back to the server
const HeartRate time.Duration = 45

// VerifyRate is how often, in seconds, the server will clean up its active host list
const VerifyRate time.Duration = 5

// KillThreshold is the limit on how many seconds a node can be unresponsive prior to killing it
const KillThreshold time.Duration = 46

// ConfigFile holds runtime configurations
const ConfigFile string = "config.json"

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

// Config holds runtime configurations loaded from common/config.json
type Config struct {
	Port  string `json:"Port"`
	Host  string `json:"Host"`
	Queue string `json:"Queue"`
}

// InvalidJobError is called when json key "type" is not "timed" "monte-carlo" or "merge-sort"
func InvalidJobError() error {
	return errors.New("invalid job")
}

// LoadConfig loads runtime configurations
func LoadConfig() (config Config) {
	filepath := path.Join("..", "src", "github.com", "johnlahut", "dist-go", "common", ConfigFile)
	file, err := os.Open(filepath)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	FailOnError(err, ">> unable to load config file")
	return
}

// MergeSort performs a merge sort on the given slice
func MergeSort(arr []float64) {

	// base case - if we have an slice of length 1
	if len(arr) > 1 {

		// compute midpoint, and create new slices (need to copy because s slice of a slice uses same memory)
		mid := len(arr) / 2
		left := make([]float64, len(arr[:mid]))
		right := make([]float64, len(arr[mid:]))
		copy(left, arr[:mid])
		copy(right, arr[mid:])

		// sort our smaller slices
		MergeSort(left)
		MergeSort(right)

		// i, j trace the sub slices
		// k traces the master slice
		i, j, k := 0, 0, 0

		// loop through placing lists in order
		for i < len(left) && j < len(right) {
			if left[i] < right[j] {
				arr[k] = left[i]
				i++
			} else {
				arr[k] = right[j]
				j++
			}
			k++
		}

		// append any left over elements
		for ; i < len(left); i++ {
			arr[k] = left[i]
			k++
		}
		for ; j < len(right); j++ {
			arr[k] = right[j]
			k++
		}
	}
}
