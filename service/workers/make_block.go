package workers

import (
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/HPISTechnologies/3rd-party/eth/common"
	ethRlp "github.com/HPISTechnologies/3rd-party/eth/rlp"
	ethTypes "github.com/HPISTechnologies/3rd-party/eth/types"
	monacoAbciTypes "github.com/HPISTechnologies/Monaco/abci/types"
	monacoCoreTypes "github.com/HPISTechnologies/Monaco/core/types"
	"github.com/HPISTechnologies/common-lib/common"
	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/log"
	"go.uber.org/zap"
)

type MakeBlock struct {
	actor.WorkerThread
}

//return a Subscriber struct
func NewMakeBlock(concurrency int, groupid string) *MakeBlock {
	in := MakeBlock{}
	in.Set(concurrency, groupid)
	return &in

}

func (m *MakeBlock) OnStart() {
}

func (m *MakeBlock) OnMessageArrived(msgs []*actor.Message) error {

	begintime0 := time.Now()
	noderole := &types.NodeRole{}
	coinbase := ethCommon.Address{}
	txhash := ethCommon.Hash{}
	accthash := ethCommon.Hash{}
	rcpthash := ethCommon.Hash{}
	gasused := uint64(0)
	txSelected := &[][]byte{}
	parentinfo := &types.ParentInfo{}
	height := uint64(0)
	vertifyheader := &types.PartialHeader{}

	currentinfo := &types.ParentInfo{}

	for _, v := range msgs {
		//m.Log.Info("received MakeBlock", "msgs", v)
		switch v.Name {
		case actor.MsgNodeRole:
			noderole = v.Data.(*types.NodeRole)
			isnil, err := m.IsNil(noderole, "noderole")
			if isnil {
				return err
			}
			m.AddLog(log.LogLevel_Info, "received in  MakeBlock", zap.String("noderole", fmt.Sprintf("%v", noderole)))
		case actor.MsgSelectedTx:
			txSelected = v.Data.(*[][]byte)
			isnil, err := m.IsNil(txSelected, "txSelected")
			if isnil {
				return err
			}
		case actor.MsgVertifyHeader:
			vertifyheader = v.Data.(*types.PartialHeader)
			isnil, err := m.IsNil(vertifyheader, "vertifyheader")
			if isnil {
				return err
			}

		case actor.MsgCoinbase:
			addr := v.Data.(*ethCommon.Address)
			isnil, err := m.IsNil(addr, "coinbase")
			if isnil {
				return err
			}
			coinbase = *addr

		case actor.MsgTxHash:
			hash := v.Data.(*ethCommon.Hash)
			isnil, err := m.IsNil(hash, "txhash")
			if isnil {
				return err
			}
			txhash = *hash
		case actor.MsgAcctHash:
			hash := v.Data.(*ethCommon.Hash)
			isnil, err := m.IsNil(hash, "accthash")
			if isnil {
				return err
			}
			accthash = *hash
		case actor.MsgRcptHash:
			hash := v.Data.(*ethCommon.Hash)
			isnil, err := m.IsNil(hash, "rcpthash")
			if isnil {
				return err
			}
			rcpthash = *hash
		case actor.MsgGasUsed:
			gas := v.Data.(uint64)
			isnil, err := m.IsNil(gas, "gasused")
			if isnil {
				return err
			}
			gasused = gas
		case actor.MsgParentInfo:
			parentinfo = v.Data.(*types.ParentInfo)
			isnil, err := m.IsNil(parentinfo, "parentinfo")
			if isnil {
				return err
			}

		case actor.MsgHeight:
			height = v.Data.(uint64)
		}

	}

	header := &ethTypes.Header{
		ParentHash: parentinfo.ParentHash,
		Number:     big.NewInt(common.Uint64ToInt64(height)),
		//GasLimit:   s.reaper.GseLimit, ??????????????????????????????????????????????????????????
		//Extra:      w.extra,
		Time:        noderole.Timestamp, //big.NewInt(time.Now().Unix()),
		Difficulty:  big.NewInt(1),
		Coinbase:    coinbase,
		Root:        accthash,
		GasUsed:     gasused,
		TxHash:      txhash,
		ReceiptHash: rcpthash,
	}

	//save cache root and header hash
	currentinfo = &types.ParentInfo{
		ParentHash: header.Hash(),
		ParentRoot: accthash,
	}

	m.AddLog(log.LogLevel_Info, "current role ", zap.String("is", noderole.Role))

	if noderole.Role == actor.MsgBlockRole_Propose {

		ethHeader, err := ethRlp.EncodeToBytes(&header)

		if err != nil {
			m.AddLog(log.LogLevel_Error, "block header eccode err", zap.String("err", err.Error()))
			return err
		}
		ExtHeader := monacoAbciTypes.ExtHeader{}
		ExtHeader.HeaderData = append(ExtHeader.HeaderData, ethHeader)
		ExtHeader.HeaderType = append(ExtHeader.HeaderType, types.TxType_Eth)

		txs := []monacoCoreTypes.Tx{}
		if txSelected != nil && len(*txSelected) > 0 {
			for i := 0; i < len(*txSelected); i++ {
				txs = append(txs, monacoCoreTypes.Tx((*txSelected)[i]))
			}
			m.AddLog(log.LogLevel_Info, "core pack", zap.Int("txsnums", len(txs)))
		}

		block := monacoCoreTypes.MakeBlock(common.Uint64ToInt64(height), txs, &monacoCoreTypes.Commit{})
		block.Header.ExtHeader = ExtHeader

		m.AddLog(log.LogLevel_Info, "core pack completed", zap.Duration("times", time.Now().Sub(begintime0)))
		m.MsgBroker.Send(actor.MsgProposeBlock, block)
		m.MsgBroker.Send(actor.MsgNewParentInfo, currentinfo)
	} else {

		code := byte(0) // all equals

		if header.Root != vertifyheader.StateRoothash {
			code = byte(1) //state roothash not equals
			m.AddLog(log.LogLevel_Error, "root is not equals", zap.String("except", fmt.Sprintf("%x", vertifyheader.StateRoothash.Bytes())), zap.String("get", fmt.Sprintf("%x", header.Root.Bytes())))
		}
		if header.ReceiptHash != vertifyheader.RcptRoothash {
			code = byte(2) //recipts roothash not equals
			m.AddLog(log.LogLevel_Error, "rcptroot is not equals", zap.String("except", fmt.Sprintf("%x", vertifyheader.RcptRoothash.Bytes())), zap.String("get", fmt.Sprintf("%x", header.ReceiptHash.Bytes())))
		}
		if header.TxHash != vertifyheader.TxRoothash {
			code = byte(3) //tx roothash not equals
			m.AddLog(log.LogLevel_Error, "txroot is not equals", zap.String("except", fmt.Sprintf("%x", vertifyheader.TxRoothash.Bytes())), zap.String("get", fmt.Sprintf("%x", header.TxHash.Bytes())))
		}

		// if code > byte(0) {
		// 	return nil
		// }

		// bv := &ntypes.BlockVertified{
		// 	Code: code,
		// 	//Msgid: vertifyheader.Msgid,
		// }
		m.MsgBroker.Send(actor.MsgBlockVertified, code)
		m.MsgBroker.Send(actor.MsgNewParentInfo, currentinfo)
	}
	return nil
}
