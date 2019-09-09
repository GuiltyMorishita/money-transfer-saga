package saga

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga/messages"
)

func NewTransferProcess(from, to *actor.PID, amount int, availability float64) *TransferProcess {
	return &TransferProcess{
		From:         from,
		To:           to,
		Amount:       amount,
		Availability: availability,
	}
}

type TransferProcess struct {
	From             *actor.PID
	To               *actor.PID
	Amount           int
	Availability     float64
	Behavior         Behavior
	Restarting       bool
	Stopping         bool
	ProcessCompleted bool
}

func (p *TransferProcess) TryCredit(targetActor *actor.PID, amount int) *actor.Props {
	return actor.FromProducer(func() actor.Actor {
		return NewAccountProxy(targetActor, func(sender *actor.PID) interface{} {
			return &messages.Credit{Amount: amount, ReplyTo: sender}
		})
	})
}

func (p *TransferProcess) TryDebit(targetActor *actor.PID, amount int) *actor.Props {
	return actor.FromProducer(func() actor.Actor {
		return NewAccountProxy(targetActor, func(sender *actor.PID) interface{} {
			return &messages.Debit{Amount: amount, ReplyTo: sender}
		})
	})
}

func (p *TransferProcess) ApplyEvent() {
}

func (p *TransferProcess) Fail() bool {
	return false
}

func (p *TransferProcess) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		ctx.Become(p.Starting)
		return
	}
}

func (p *TransferProcess) Starting(ctx actor.Context) {
	if _, ok := ctx.Message().(*actor.Started); ok {
		ctx.SpawnNamed(p.TryDebit(p.From, -p.Amount), "DebitAttempt")
	}
}
