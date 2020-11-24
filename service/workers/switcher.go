package workers

import (
	"github.com/HPISTechnologies/component-lib/actor"
)

type Switcher struct {
	actor.WorkerThread
}

//return a Subscriber struct
func NewSwitcher(concurrency int, groupid string) *Switcher {
	switcher := Switcher{}
	switcher.Set(concurrency, groupid)
	return &switcher
}

func (switcher *Switcher) OnStart() {
}

func (switcher *Switcher) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgInitParentInfo:
			switcher.MsgBroker.Send(actor.MsgParentInfo, v.Data)
		}
	}
	return nil
}
