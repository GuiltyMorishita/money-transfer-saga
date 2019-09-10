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
			NewTransferFactory()
		}

	case messages.SuccessResult:
	case messages.UnknownResult:
	case messages.FailedAndInconsistent:
	case messages.FailedButConsistentResult:
	}
}

func NewTransferFactory(
	availability float64,
	retryAttempts int,
	ctx actor.Context,
) *TransferFactory {
	return &TransferFactory{
		availability:  availability,
		retryAttempts: retryAttempts,
		ctx:           ctx,
	}
}

type TransferFactory struct {
	availability  float64
	retryAttempts int
	ctx           actor.Context
}

func (f *TransferFactory) CreateTransfer(actorName string, from, to *actor.PID, amount int) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return saga.NewTransferProcess(from, to, amount, f.availability)
	}).WithSupervisor(
		actor.NewOneForOneStrategy(f.retryAttempts, 1000, actor.DefaultDecider),
	)
	return actor.SpawnNamed(props, actorName)
}
