package messages

import "github.com/AsynkronIT/protoactor-go/actor"

type Result struct {
	Pid *actor.PID
}

type UnknownResult struct {
	Pid *actor.PID
}

type FailedButConsistentResult struct {
	Pid *actor.PID
}

type FailedAndInconsistent struct {
	Pid *actor.PID
}

type SuccessResult struct {
	Pid *actor.PID
}
