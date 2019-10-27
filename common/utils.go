package common

import "errors"

// LocalQueue - used for tests
const LocalQueue string = "amqp://guest:guest@localhost:5672/"

// TimedJobType - json key for timed jobs
const TimedJobType = "timed"

// MonteCarloJobType - json key for monte-carlo jobs
const MonteCarloJobType = "monte-carlo"

// MergeSortJobType - json key for merge-sort jobs
const MergeSortJobType = "merge-sort"

// DevPort for server while in development
const DevPort string = ":8090"

// Job represents an incoming job to the main server
type Job struct {
	ID   int
	Type string   `json:"type"`
	Data []uint16 `json:"data"`
}

// TimedJob represents a job to trigger the process to wait for "time"
type TimedJob struct {
	Time int `json:"time"`
}

// InvalidJobError is called when json key "type" is not "timed" "monte-carlo" or "merge-sort"
func InvalidJobError() error {
	return errors.New("invalid job")
}
