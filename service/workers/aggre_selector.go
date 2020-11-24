package workers

import (
	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/aggregator/aggregator"
	"github.com/HPISTechnologies/component-lib/log"
	"go.uber.org/zap"
)

type AggreSelector struct {
	actor.WorkerThread

	savedMessage *actor.Message
	aggregator   *aggregator.Aggregator

	counter int
}

//return a Subscriber struct
func NewAggreSelector(concurrency int, groupid string) *AggreSelector {
	agg := AggreSelector{}
	agg.Set(concurrency, groupid)
	agg.aggregator = aggregator.NewAggregator()

	return &agg
}

func (a *AggreSelector) OnStart() {
}

func (a *AggreSelector) OnMessageArrived(msgs []*actor.Message) error {
	switch msgs[0].Name {
	case actor.MsgBlockCompleted:
		result := msgs[0].Data.(string)
		if result == actor.MsgBlockCompleted_Success {
			remainingQuantity := a.aggregator.OnClearInfoReceived()
			a.AddLog(log.LogLevel_Debug, "core AggreSelector clear pool", zap.Int("remainingQuantity", remainingQuantity))
		}

	case actor.MsgFinalInclusive:
		inclusive := msgs[0].Data.(*types.InclusiveList)
		a.savedMessage = msgs[0].CopyHeader()
		inclusive.Mode = types.InclusiveMode_Message
		result, missingSize := a.aggregator.OnListReceived(inclusive)
		a.AddLog(log.LogLevel_Debug, "core AggreSelector exec MsgInclusive", zap.Int("missingSize", missingSize), zap.Int("received counter", a.counter))
		a.counter = 0
		a.SendMsg(result, a.MsgBroker)

	case actor.MsgRawtx:
		standardTransactions := msgs[0].Data.([]*types.StandardTransaction)
		for i := range standardTransactions {
			result := a.aggregator.OnDataReceived(standardTransactions[i].TxHash, standardTransactions[i].TxRawData)
			a.SendMsg(result, a.MsgBroker)
		}
		a.counter = a.counter + len(standardTransactions)
	}

	return nil
}
func (a *AggreSelector) SendMsg(selectedData *[]*interface{}, sender *actor.MessageWrapper) {
	a.ChangeEnvironment(a.savedMessage)

	if selectedData != nil {
		txs := make([][]byte, len(*selectedData))
		for i, tx := range *selectedData {
			txs[i] = (*tx).([]byte)
		}

		a.AddLog(log.LogLevel_CheckPoint, "aggreSelector execute complete", zap.Int("PackNums", len(txs)))
		sender.Send(actor.MsgSelectedTx, &txs)
	}
}
