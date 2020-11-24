package workers

import (
	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/actor"
)

type Cache struct {
	actor.WorkerThread
	parentinfo *types.ParentInfo
}

//return a Subscriber struct
func NewCache(concurrency int, groupid string) *Cache {
	cache := Cache{}
	cache.Set(concurrency, groupid)
	cache.parentinfo = &types.ParentInfo{}
	return &cache
}

func (caches *Cache) OnStart() {
}

func (cache *Cache) OnMessageArrived(msgs []*actor.Message) error {
	var newparentinfo *types.ParentInfo
	result := ""

	for _, v := range msgs {

		switch v.Name {
		case actor.MsgBlockCompleted:
			result = v.Data.(string)
		case actor.MsgNewParentInfo:
			newparentinfo = v.Data.(*types.ParentInfo)
			isnil, err := cache.IsNil(newparentinfo, "parentinfo")
			if isnil {
				return err
			}
		}
	}

	if actor.MsgBlockCompleted_Failed == result {
		cache.MsgBroker.Send(actor.MsgParentInfo, cache.parentinfo)
	} else if actor.MsgBlockCompleted_Success == result && newparentinfo != nil {
		cache.MsgBroker.Send(actor.MsgParentInfo, newparentinfo)
		cache.parentinfo = newparentinfo
	}
	return nil
}
