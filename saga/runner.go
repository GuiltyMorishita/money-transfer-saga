package saga

import (
	"fmt"

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
	// private RootContext Context = RootContext.Empty;
	intervalBetweenConsoleUpdates int
	numberOfIterations            int
	uptime                        float64
	refusalProbability            float64
	busyProbability               float64
	retryAttempts                 int
	verbose                       bool
	// private readonly HashSet<PID> _transfers = new HashSet<PID>();
	successResults               int
	failedAndInconsistentResults int
	failedButConsistentResults   int
	unknownResults               int
}

func (r *Runner) CreateAccount(name string) *actor.PID {
	props := actor.FromProducer(func() actor.Actor { return NewAccount(name, r.uptime, r.refusalProbability, r.busyProbability) })
	return actor.SpawnNamed(props, name)
}

func (r *Runner) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		for i := 0; i < r.numberOfIterations; i++ {
			r.CreateAccount(fmt.Sprintf("FromAccount{%d}", i))
			r.CreateAccount(fmt.Sprintf("ToAccount{%d}", i))
		}

	case messages.SuccessResult:
	case messages.UnknownResult:
	case messages.FailedAndInconsistent:
	case messages.FailedButConsistentResult:
	}
}
