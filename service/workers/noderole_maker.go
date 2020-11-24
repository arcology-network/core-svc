package workers

import (
	"fmt"
	"math/big"
	"time"

	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type NodeRoleMaker struct {
	actor.WorkerThread
}

//return a Subscriber struct
func NewNodeRoleMaker(concurrency int, groupid string) *NodeRoleMaker {
	maker := NodeRoleMaker{}
	maker.Set(concurrency, groupid)
	return &maker
}

func (maker *NodeRoleMaker) OnStart() {
}

func (maker *NodeRoleMaker) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgNodeRole:
			nodeRole := v.Data.(*types.NodeRole)

			//nodeRole.Timestamp = decimal.NewFromInt(viper.GetInt64("starter")).Add(decimal.NewFromInt(time.Now().Unix()).Mul(decimal.NewFromFloat(viper.GetFloat64("rate")))).BigInt()
			multiResult := big.NewInt(0).Mul(big.NewInt(time.Now().Unix()), big.NewInt(viper.GetInt64("rate")))
			nodeRole.Timestamp = big.NewInt(0).Add(big.NewInt(viper.GetInt64("starter")), multiResult)
			maker.AddLog(log.LogLevel_Info, " node role maker", zap.String("noderole", fmt.Sprintf("%v", nodeRole)))

			maker.MsgBroker.Send(actor.MsgNodeRoleCore, nodeRole)
		}
	}
	return nil
}
