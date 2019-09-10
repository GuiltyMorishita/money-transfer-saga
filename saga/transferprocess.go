package saga

import (
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga/messages"
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
		return NewTransferProcess(from, to, amount, f.availability)
	}).WithSupervisor(
		actor.NewOneForOneStrategy(f.retryAttempts, 1000, actor.DefaultDecider),
	)
	return f.ctx.SpawnNamed(props, actorName)
}

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
			return
		}

	case *actor.Terminated:
		if p.Restarting || p.Stopping {
			// if the TransferProcess itself is restarting or stopping due to failure, we will receive a
			// Terminated message for any child actors due to them being stopped but we should not
			// treat this as a failure of the saga, so return here to stop further processing
		}
		return
	}

	ctx.Self().RequestFuture(ctx.Message(), 10*time.Second).Result()
}

func (p *TransferProcess) Starting(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		log.Println("Try Debit")
		ctx.Become(p.AwaitingDebitConfirmation)
		ctx.SpawnNamed(p.TryDebit(p.From, -p.Amount), "DebitAttempt")
	}
}

func (p *TransferProcess) AwaitingDebitConfirmation(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryDebit(p.From, -p.Amount), "DebitAttempt")

	case *messages.OK:
		log.Println("Try Credit")
		ctx.Become(p.AwaitingCreditConfirmation)
		ctx.SpawnNamed(p.TryCredit(p.To, +p.Amount), "CreditAttempt")

	case *messages.Refused:
		log.Println("Debit refused. System consistent")
		p.ProcessCompleted = true
		ctx.Parent().Tell(messages.FailedButConsistentResult{Pid: ctx.Self()})
		p.StopAll(ctx)

	case *actor.Terminated:
		log.Println("Debit status unknown. Escalate")
		p.ProcessCompleted = true
		p.StopAll(ctx)
	}
}

func (p *TransferProcess) AwaitingCreditConfirmation(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryCredit(p.To, +p.Amount), "CreditAttempt")

	case *messages.OK:
		log.Println("Success!")
		p.ProcessCompleted = true
		ctx.Parent().Tell(messages.SuccessResult{Pid: ctx.Self()})
		p.StopAll(ctx)

	case *messages.Refused:
		ctx.Become(p.RollingBackDebit)
		ctx.SpawnNamed(p.TryCredit(p.From, +p.Amount), "RollbackDebit")

	case *actor.Terminated:
		log.Println("Credit status unknown. Escalate")
		p.ProcessCompleted = true
		p.StopAll(ctx)
	}
}

func (p *TransferProcess) RollingBackDebit(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.SpawnNamed(p.TryCredit(p.From, +p.Amount), "RollbackDebit")

	case *messages.OK:
		log.Println("Transfer failed. System consistent")
		p.ProcessCompleted = true
		ctx.Parent().Tell(messages.FailedButConsistentResult{Pid: ctx.Self()})
		p.StopAll(ctx)

	case *messages.Refused, *actor.Terminated:
		log.Println("Transfer status unknown. Escalate")
		ctx.Parent().Tell(messages.FailedAndInconsistent{Pid: ctx.Self()})
		p.StopAll(ctx)
	}
}

func (p *TransferProcess) StopAll(ctx actor.Context) {
	p.From.Stop()
	p.To.Stop()
	ctx.Self().Stop()
}
