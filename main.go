package main

import (
	"log"

	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga"
)

func main() {
	var (
		numberOfTransfers  = 5
		uptime             = 99.99
		refusalProbability = 0.01
		busyProbability    = 0.01
		retryAttempts      = 0
		verbose            = false
	)

	log.Println("Starting")
	props := actor.FromProducer(func() actor.Actor {
		return saga.NewRunner(numberOfTransfers, uptime, refusalProbability, busyProbability, retryAttempts, verbose)
	})

	log.Println("Spawning runner")
	actor.SpawnNamed(props, "runner")

	console.ReadLine()
}
