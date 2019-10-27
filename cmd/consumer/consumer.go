package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/johnlahut/dist-go/common"
)

// Start will start consuming and processing from the queue
func main() {
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

	forever := make(chan bool)

	// listen forever
	go func() {
		for d := range msgs {

			// unmarshal job
			var job common.Job
			err := json.Unmarshal(d.Body, &job)
			common.FailOnError(err, "failed to unmarshal job")

			// check job type
			switch job.Type {
			case common.TimedJobType:
				log.Printf("received timed job: %s", d.Body)

				t := time.Duration(job.Data[0])
				time.Sleep(t * time.Second)
			case common.MonteCarloJobType:
				log.Printf("received monte-carlo job: %s", d.Body)
			case common.MergeSortJobType:
				log.Printf("received merge-sort job: %s", d.Body)
			default:
				common.FailOnError(common.InvalidJobError(), "invalid job type")
			}

			// positive ack
			d.Ack(false)
		}
	}()

	log.Printf(" [*] waiting for messages. to exit press ctrl-c")
	<-forever
}
