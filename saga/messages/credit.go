package messages

import "github.com/AsynkronIT/protoactor-go/actor"

type Credit struct {
	ReplyTo *actor.PID
	Amount  int
}

func (c *Credit) ChangeBalance(replyTo *actor.PID, amount int) {
	c.ReplyTo = replyTo
	c.Amount = amount
}
