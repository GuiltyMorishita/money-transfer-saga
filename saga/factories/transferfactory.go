package factories

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga"
)

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
