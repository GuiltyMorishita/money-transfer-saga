package saga

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/GuiltyMorishita/money-transfer-saga/saga/messages"
)

func NewAccountProxy(target *actor.PID, createMessage func(*actor.PID) interface{}) *AccountProxy {
	return &AccountProxy{
		Target:        target,
		CreateMessage: createMessage,
	}
}

type AccountProxy struct {
	Target        *actor.PID
	CreateMessage func(*actor.PID) interface{}
}

func (p *AccountProxy) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		p.Target.Tell(p.CreateMessage(ctx.Self()))
		ctx.SetReceiveTimeout(100 * time.Millisecond)
		return

	case *messages.OK:
		ctx.CancelReceiveTimeout()
		ctx.Parent().Tell(msg)
		return

	case *messages.Refused:
		ctx.CancelReceiveTimeout()
		ctx.Parent().Tell(msg)
		return

	case *messages.InsufficientFunds:
	case *messages.InternalServerError:
	case *messages.ServiceUnavailable:
	case *actor.ReceiveTimeout:
	}
}
