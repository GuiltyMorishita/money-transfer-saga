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

func (p *TransferProcess) TryCredit(targetActor *actor.PID, amount int) actor.Props {
	return actor.FromProducer(func() actor.Actor {
		return NewAccountProxy(targetActor, func(sender *actor.PID) interface{} {
			return &messages.Credit{Amount: amount, ReplyTo: sender}
		})
	})
}

func (p *TransferProcess) TryDebit(targetActor *actor.PID, amount int) actor.Props {
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
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.Become(p.Starting)

	case *actor.Stopping:
		p.Stopping = true

	case *actor.Restarting:
		p.Restarting = true

	case *actor.Stopped:
		if !p.ProcessCompleted {
			ctx.Parent().Tell(messages.UnknownResult{Pid: ctx.Self()})
		}

	case *actor.Terminated:
		if p.Restarting || p.Stopping {
			// if the TransferProcess itself is restarting or stopping due to failure, we will receive a
			// Terminated message for any child actors due to them being stopped but we should not
			// treat this as a failure of the saga, so return here to stop further processing
		}
	}
	ctx.Receive(ctx)
}

func (p *TransferProcess) Starting(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryDebit(p.From, -p.Amount), "DebitAttempt")
	}
}

func (p *TransferProcess) AwaitingDebitConfirmation(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryDebit(p.From, -p.Amount), "DebitAttempt")

	case *messages.OK:
		ctx.SpawnNamed(p.TryCredit(p.From, +p.Amount), "CreditAttempt")

	case *messages.Refused:
		ctx.Parent().Tell(messages.FailedButConsistentResult{Pid: ctx.Self()})
		p.StopAll(ctx)

	case *actor.Terminated:
		p.StopAll(ctx)
	}
}

func (p *TransferProcess) AwaitingCreditConfirmation(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryCredit(p.To, +p.Amount), "CreditAttempt")

	case *messages.OK:
		ctx.Parent().Tell(messages.SuccessResult{Pid: ctx.Self()})
		p.StopAll(ctx)

	case *messages.Refused:
		ctx.SpawnNamed(p.TryCredit(p.From, +p.Amount), "RollbackDebit")

	case *actor.Terminated:
		p.StopAll(ctx)
	}
}

func (p *TransferProcess) RollingBackDebit(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryCredit(p.From, +p.Amount), "RollbackDebit")

	case *messages.OK:
		ctx.Parent().Tell(messages.FailedButConsistentResult{Pid: ctx.Self()})
		p.StopAll(ctx)

	case *messages.Refused, *actor.Terminated:
		ctx.Parent().Tell(messages.FailedAndInconsistent{Pid: ctx.Self()})
		p.StopAll(ctx)
	}
}

func (p *TransferProcess) StopAll(ctx actor.Context) {
	p.From.Stop()
	p.To.Stop()
	ctx.Self().Stop()
}
