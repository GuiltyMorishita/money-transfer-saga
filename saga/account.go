package saga

import (
	"github.com/AsynkronIT/protoactor-go/actor"
)

func NewAccount(name string, serviceUptime, refusalProbability, busyProbability int64, random int) *Account {
	return &Account{
		Name:               name,
		ServiceUptime:      serviceUptime,
		refusalProbability: refusalProbability,
		busyProbability:    busyProbability,
		random:             random,
	}
}

type Account struct {
	Name               string
	ServiceUptime      int64
	refusalProbability int64
	busyProbability    int64
	processedMessages  map[*actor.PID]interface{}
	balance            int
	random             int
}
