package workers

import (
	"fmt"

	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/log"
	"go.uber.org/zap"
)

type MakeBlockPreparation struct {
	actor.WorkerThread
}

//return a Subscriber struct
func NewMakeBlockPreparation(concurrency int, groupid string) *MakeBlockPreparation {

	mbp := MakeBlockPreparation{}
	mbp.Set(concurrency, groupid)

	return &mbp
}

func (m *MakeBlockPreparation) OnStart() {
}

func (m *MakeBlockPreparation) OnMessageArrived(msgs []*actor.Message) error {

	for _, v := range msgs {

		switch v.Name {
		case actor.MsgNodeRole:
			nodeRole := v.Data.(*types.NodeRole)
			m.AddLog(log.LogLevel_Info, " MakeBlock preparation", zap.String("noderole", fmt.Sprintf("%v", nodeRole)))
			if actor.MsgBlockRole_Propose == nodeRole.Role {
				nilHeader := types.PartialHeader{}
				m.MsgBroker.Send(actor.MsgVertifyHeader, &nilHeader)
			}

		}
	}

	return nil
}
