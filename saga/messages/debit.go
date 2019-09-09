package messages

import "github.com/AsynkronIT/protoactor-go/actor"

type Debit struct {
	ReplyTo *actor.PID
	Amount  int
}

func (c *Debit) ChangeBalance(replyTo *actor.PID, amount int) {
	c.ReplyTo = replyTo
	c.Amount = amount
}
