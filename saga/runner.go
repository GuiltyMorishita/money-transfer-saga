package saga

import (
	"fmt"
	"log"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga/messages"
)

func NewRunner(
	numberOfIterations int,
	uptime float64,
	refusalProbability float64,
	busyProbability float64,
	retryAttempts int,
	verbose bool,
) *Runner {
	return &Runner{
		numberOfIterations: numberOfIterations,
		uptime:             uptime,
		refusalProbability: refusalProbability,
		busyProbability:    busyProbability,
		retryAttempts:      retryAttempts,
		verbose:            verbose,
	}
}

type Runner struct {
	intervalBetweenConsoleUpdates int
	numberOfIterations            int
	uptime                        float64
	refusalProbability            float64
	busyProbability               float64
	retryAttempts                 int
	verbose                       bool
	successResults                int
	failedAndInconsistentResults  int
	failedButConsistentResults    int
	unknownResults                int
}

func (r *Runner) CreateAccount(name string) *actor.PID {
	props := actor.FromProducer(func() actor.Actor { return NewAccount(name, r.uptime, r.refusalProbability, r.busyProbability) })
	return actor.SpawnNamed(props, name)
}

func (r *Runner) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		for i := 0; i < r.numberOfIterations; i++ {
			from := r.CreateAccount(fmt.Sprintf("FromAccount %d", i))
			to := r.CreateAccount(fmt.Sprintf("ToAccount %d", i))
			factory := NewTransferFactory(r.uptime, r.retryAttempts, ctx)
			factory.CreateTransfer(fmt.Sprintf("Transfer Prossess %d", i), from, to, 10)
			log.Println(fmt.Sprintf("Started %d / %d", i+1, r.numberOfIterations))
		}

	case *messages.SuccessResult:
		r.successResults++
		log.Println("successResult")

	case *messages.UnknownResult:
		r.unknownResults++
	case *messages.FailedAndInconsistent:
		r.failedAndInconsistentResults++
	case *messages.FailedButConsistentResult:
		r.failedButConsistentResults++
	}
}
