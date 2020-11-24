package workers

import (
	ethCommon "github.com/HPISTechnologies/3rd-party/eth/common"
	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/actor"
)

type InclusizeMaker struct {
	actor.WorkerThread
}

//return a Subscriber struct
func NewInclusizeMaker(concurrency int, groupid string) *InclusizeMaker {

	mbp := InclusizeMaker{}
	mbp.Set(concurrency, groupid)

	return &mbp
}

func (m *InclusizeMaker) OnStart() {
}

func (m *InclusizeMaker) OnMessageArrived(msgs []*actor.Message) error {

	var inclusive *types.InclusiveList
	var finalListSpawn *[]*ethCommon.Hash

	for _, v := range msgs {
		switch v.Name {
		case actor.MsgInclusive:
			inclusive = msgs[0].Data.(*types.InclusiveList)

		case actor.MsgFinalTxsListSpawned:
			finalListSpawn = v.Data.(*[]*ethCommon.Hash)
		}

	}

	if inclusive != nil && finalListSpawn != nil {
		removes := map[ethCommon.Hash]int{}
		for i, hash := range *finalListSpawn {
			removes[*hash] = i
		}
		totalLength := len(inclusive.Successful)
		listHash := make([]*ethCommon.Hash, 0, totalLength)
		listFlag := make([]bool, 0, totalLength)

		for i := range inclusive.Successful {
			if _, ok := removes[*inclusive.HashList[i]]; !ok {
				listHash = append(listHash, inclusive.HashList[i])
				listFlag = append(listFlag, inclusive.Successful[i])
			}
		}
		inclusive.HashList = listHash
		inclusive.Successful = listFlag
	}

	m.MsgBroker.Send(actor.MsgFinalInclusive, inclusive)

	return nil
}
