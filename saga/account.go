package saga

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga/messages"
)

func NewAccount(name string, serviceUptime, refusalProbability, busyProbability int64) *Account {
	return &Account{
		Name:               name,
		ServiceUptime:      serviceUptime,
		RefusalProbability: refusalProbability,
		BusyProbability:    busyProbability,
	}
}

type Account struct {
	Name               string
	ServiceUptime      int64
	RefusalProbability int64
	BusyProbability    int64
	ProcessedMessages  map[*actor.PID]interface{}
	Balance            int
}

func (a *Account) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case messages.Credit:
		if a.AlreadyProcessed(msg.ReplyTo) {
			a.Reply(msg.ReplyTo)
			return
		}
		a.AdjustBalance(msg.ReplyTo, msg.Amount)

	case messages.Debit:
		if a.AlreadyProcessed(msg.ReplyTo) {
			a.Reply(msg.ReplyTo)
			return
		}

		if msg.Amount+a.Balance >= 0 {
			a.AdjustBalance(msg.ReplyTo, msg.Amount)
			return
		}

		msg.ReplyTo.Tell(messages.InsufficientFunds{})

	case messages.GetBalance:
		ctx.Respond(a.Balance)
	}
}

func (a *Account) Reply(replyTo *actor.PID) {
	replyTo.Tell(a.ProcessedMessages[replyTo])
}

func (a *Account) AdjustBalance(replyTo *actor.PID, amount int) {
	if a.RefusePermanently() {
		a.ProcessedMessages[replyTo] = messages.Refused{}
		replyTo.Tell(messages.Refused{})
		return
	}

	if a.Busy() {
		replyTo.Tell(messages.ServiceUnavailable{})
	}

	behavior := a.DetermineProcessingBehavior()
	if behavior == FailBeforeProcessing {
		a.Failure(replyTo)
		return
	}

	time.Sleep(150 * time.Millisecond)

	a.Balance += amount
	a.ProcessedMessages[replyTo] = messages.OK{}

	if behavior == FailAfterProcessing {
		a.Failure(replyTo)
		return
	}

	replyTo.Tell(messages.OK{})
}

func (a *Account) Busy() bool {
	return false
}

func (a *Account) RefusePermanently() bool {
	return false
}

func (a *Account) Failure(replyTo *actor.PID) {
	replyTo.Tell(messages.InternalServerError{})
}

func (a *Account) DetermineProcessingBehavior() Behavior {
	return ProcessSuccessfully
}

func (a *Account) AlreadyProcessed(replyTo *actor.PID) bool {
	_, ok := a.ProcessedMessages[replyTo]
	return ok
}

type Behavior int

const (
	FailBeforeProcessing = iota
	FailAfterProcessing
	ProcessSuccessfully
)
