package workers

import (
	"fmt"
	"time"

	ethCommon "github.com/HPISTechnologies/3rd-party/eth/common"
	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/log"
	"github.com/HPISTechnologies/component-lib/mhasher"
	"go.uber.org/zap"
)

type CalculateTxHash struct {
	actor.WorkerThread
}

//return a Subscriber struct
func NewCalculateTxHash(concurrency int, groupid string) *CalculateTxHash {
	in := CalculateTxHash{}
	in.Set(concurrency, groupid)

	return &in
}

func (c *CalculateTxHash) OnStart() {
}

func (c *CalculateTxHash) OnMessageArrived(msgs []*actor.Message) error {
	txs := msgs[0].Data.(*[][]byte)
	roothash := ethCommon.Hash{}
	if txs != nil && len(*txs) > 0 {
		begintime1 := time.Now()
		roothash = mhasher.GetTxsHash(*txs)

		c.AddLog(log.LogLevel_Debug, "calculate txroothash", zap.Duration("times", time.Now().Sub(begintime1)))
	}

	c.AddLog(log.LogLevel_CheckPoint, "send txroothash", zap.String("roothash", fmt.Sprintf("%x", roothash)))
	c.MsgBroker.Send(actor.MsgTxHash, &roothash)
	return nil
}
