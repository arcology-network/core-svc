package service

import (
	"net/http"

	"github.com/HPISTechnologies/component-lib/actor"
	"github.com/HPISTechnologies/component-lib/streamer"
	"github.com/HPISTechnologies/core-svc/service/workers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	// "github.com/sirupsen/logrus"
	"github.com/HPISTechnologies/component-lib/kafka"
)

type Config struct {
	concurrency int
	groupid     string
}

//return a Subscriber struct
func NewConfig() *Config {
	return &Config{
		concurrency: viper.GetInt("concurrency"),
		groupid:     "core",
	}
}

func (cfg *Config) Start() {

	// logrus.SetLevel(logrus.DebugLevel)
	// logrus.SetFormatter(&logrus.JSONFormatter{})

	http.Handle("/streamer", promhttp.Handler())
	go http.ListenAndServe(":19003", nil)

	broker := streamer.NewStatefulStreamer()

	//00 initializer
	initializer := actor.NewActor(
		"initializer",
		broker,
		[]string{actor.MsgStarting},
		[]string{actor.MsgStartSub},
		[]int{1},
		workers.NewInitializer(cfg.concurrency, cfg.groupid),
	)
	initializer.Connect(streamer.NewDisjunctions(initializer, 1))

	receiveMseeages := []string{
		actor.MsgInclusive,
		actor.MsgTxHash,
		actor.MsgRcptHash,
		actor.MsgAcctHash,
		actor.MsgCoinbase,
		actor.MsgGasUsed,
		actor.MsgHeight,
		actor.MsgBlockCompleted,
		actor.MsgNodeRole,
		actor.MsgVertifyHeader,
		actor.MsgInitParentInfo,
		actor.MsgRawtx,
		actor.MsgFinalTxsListSpawned,
	}

	msgexch := viper.GetString("msgexch")

	receiveTopics := []string{
		viper.GetString("inclusive-txs"),
		msgexch,
		viper.GetString("partial-head"),
		viper.GetString("raw-txs"),
		viper.GetString("txs-list-spawned"),
	}

	//01 download
	kafkaDownloader := actor.NewActor(
		"kafkaDownloader",
		broker,
		[]string{actor.MsgStartSub},
		receiveMseeages,
		[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2000000, 1},
		kafka.NewKafkaDownloader(cfg.concurrency, cfg.groupid, receiveTopics, receiveMseeages),
	)
	kafkaDownloader.Connect(streamer.NewDisjunctions(kafkaDownloader, 2000000))

	//01 inclusiveMaker
	inclusiveMaker := actor.NewActor(
		"inclusiveMaker",
		broker,
		[]string{
			actor.MsgInclusive,
			actor.MsgFinalTxsListSpawned,
		},
		[]string{
			actor.MsgFinalInclusive,
		},
		[]int{1},
		workers.NewInclusizeMaker(cfg.concurrency, cfg.groupid),
	)
	inclusiveMaker.Connect(streamer.NewConjunctions(inclusiveMaker))

	//02 aggre
	aggreSelector := actor.NewActor(
		"aggreSelector",
		broker,
		[]string{
			actor.MsgRawtx,
			actor.MsgBlockCompleted,
			actor.MsgFinalInclusive},
		[]string{actor.MsgSelectedTx},
		[]int{1},
		workers.NewAggreSelector(cfg.concurrency, cfg.groupid),
	)
	aggreSelector.Connect(streamer.NewDisjunctions(aggreSelector, 10000))

	//03 calculateTxhash
	calculateTxHash := actor.NewActor(
		"calculateTxHash",
		broker,
		[]string{actor.MsgSelectedTx},
		[]string{actor.MsgTxHash},
		[]int{1},
		workers.NewCalculateTxHash(cfg.concurrency, cfg.groupid),
	)
	calculateTxHash.Connect(streamer.NewDisjunctions(calculateTxHash, 1))
	//04 make block prepare
	makeBlockPreparation := actor.NewActor(
		"makeBlockPreparation",
		broker,
		[]string{
			actor.MsgNodeRole,
		},
		[]string{actor.MsgVertifyHeader},
		[]int{1},
		workers.NewMakeBlockPreparation(cfg.concurrency, cfg.groupid),
	)
	makeBlockPreparation.Connect(streamer.NewDisjunctions(makeBlockPreparation, 1))

	//05 make block
	makeBlock := actor.NewActor(
		"makeBlock",
		broker,
		[]string{
			//actor.MsgBlockPreparation,
			actor.MsgNodeRole,
			actor.MsgSelectedTx,
			actor.MsgVertifyHeader,
			actor.MsgCoinbase,
			actor.MsgTxHash,
			actor.MsgRcptHash,
			actor.MsgAcctHash,
			actor.MsgGasUsed,
			actor.MsgHeight,
			actor.MsgParentInfo,
		},
		[]string{
			actor.MsgProposeBlock,
			actor.MsgBlockVertified,
			actor.MsgNewParentInfo,
		},
		[]int{1, 1, 1, 1},
		workers.NewMakeBlock(cfg.concurrency, cfg.groupid),
	)
	makeBlock.Connect(streamer.NewConjunctions(makeBlock))

	//06 cache
	cache := actor.NewActor(
		"InfoCache",
		broker,
		[]string{
			actor.MsgNewParentInfo,
			actor.MsgBlockCompleted,
		},
		[]string{
			actor.MsgParentInfo,
		},
		[]int{1},
		workers.NewCache(cfg.concurrency, cfg.groupid),
	)
	cache.Connect(streamer.NewConjunctions(cache))

	//07 switcher
	switcher := actor.NewActor(
		"switcher",
		broker,
		[]string{
			actor.MsgInitParentInfo,
		},
		[]string{
			actor.MsgParentInfo,
		},
		[]int{1},
		workers.NewSwitcher(cfg.concurrency, cfg.groupid),
	)
	switcher.Connect(streamer.NewDisjunctions(switcher, 1))

	relations := map[string]string{}
	relations[actor.MsgBlockVertified] = msgexch
	relations[actor.MsgProposeBlock] = viper.GetString("local-block")
	relations[actor.MsgParentInfo] = msgexch
	relations[actor.MsgNodeRoleCore] = msgexch
	//06 pub
	kafkaUploader := actor.NewActor(
		"kafkaUploader",
		broker,
		[]string{
			actor.MsgBlockVertified,
			actor.MsgProposeBlock,
			actor.MsgParentInfo,
			actor.MsgNodeRoleCore,
		},
		[]string{},
		[]int{},
		kafka.NewKafkaUploader(cfg.concurrency, cfg.groupid, relations),
	)
	kafkaUploader.Connect(streamer.NewDisjunctions(kafkaUploader, 4))

	//07 noderolemaker
	nodeRoleMaker := actor.NewActor(
		"nodeRoleMaker",
		broker,
		[]string{
			actor.MsgNodeRole,
		},
		[]string{
			actor.MsgNodeRoleCore,
		},
		[]int{1},
		workers.NewNodeRoleMaker(cfg.concurrency, cfg.groupid),
	)
	nodeRoleMaker.Connect(streamer.NewDisjunctions(nodeRoleMaker, 4))

	//starter
	selfStarter := streamer.NewDefaultProducer("selfStarter", []string{actor.MsgStarting}, []int{1})
	broker.RegisterProducer(selfStarter)

	broker.Serve()

	//start signal
	streamerStarting := actor.Message{
		Name:   actor.MsgStarting,
		Height: 0,
		Round:  0,
		Data:   "start",
	}
	broker.Send(actor.MsgStarting, &streamerStarting)

}

func (cfg *Config) Stop() {

}
