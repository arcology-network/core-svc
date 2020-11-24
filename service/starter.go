package service

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmCommon "github.com/HPISTechnologies/3rd-party/tm/common"

	"github.com/HPISTechnologies/component-lib/log"
)

// var (
// 	//config = cfg.DefaultConfig()
// 	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
// )

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start core service Daemon",
	RunE:  startCmd,
}

func init() {
	flags := StartCmd.Flags()

	flags.String("mqaddr", "localhost:9092", "host:port of kafka ")

	flags.String("local-block", "local-block", "topic of received proposer block ")

	flags.String("msgexch", "msgexch", "topic of received or send msg exchange")

	flags.String("raw-txs", "raw-txs", "topic of received txs")

	flags.String("inclusive-txs", "inclusive-txs", "topic of received txlist")

	flags.String("log", "log", "topic for send log")

	flags.Int("concurrency", 4, "num of threads")

	flags.String("partial-head", "partial-head", "topic of send validate block header ")

	flags.String("logcfg", "./log.toml", "log conf path")

	flags.Uint64("svcid", 10, "service id of core,range 1 to 255")
	flags.Uint64("insid", 1, "instance id of core,range 1 to 255")

	flags.Bool("draw", false, "draw flow graph")

	flags.String("txs-list-spawned", "txs-list-spawned", "topic of received  spawned txs list")

	flags.Int64("rate", 1, "the rate of timestamp")

	flags.Int64("starter", 0, "the starter of timestamp")

	flags.Int("nidx", 0, "node index in cluster")
	flags.String("nname", "node1", "node name in cluster")
}

func startCmd(cmd *cobra.Command, args []string) error {

	log.InitLog("core.log", viper.GetString("logcfg"), "core", viper.GetString("nname"), viper.GetInt("nidx"))
	//logger = logger.With("svc", "core")

	m := NewConfig()

	startSvc(m)

	return nil
}

func startSvc(bs *Config) error {
	bs.Start()

	if viper.GetBool("draw") {
		log.CompleteMetaInfo("core")
		return nil
	}

	// Wait forever
	tmCommon.TrapSignal(func() {
		// Cleanup
		bs.Stop()
	})
	return nil

}
