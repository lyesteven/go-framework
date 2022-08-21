package main

import (
	"flag"
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"github.com/nats-io/go-nats-streaming"
	log "github.com/xiaomi-tc/log15"
	"os"
	"os/signal"
	//"strconv"
	//"time"
)

var (
	i           int
	subC        chan os.Signal
	cleanupDone chan bool
)

var Msgcb stan.MsgHandler = func(msg *stan.Msg) {
	log.Info("Received trans msg.", "msg", msg.Data)
	//订单已经支付成功，开始打印发票...
}

//接收事务性消息跟其它普通消息一样，只需指定 topic,group 和一个处理消息的回调函数即可。
//在回调函数内开始打印发票
func main() {
	var startSeq uint64
	var clientID, topic string

	flag.StringVar(&clientID, "id", "TransTopicID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic, "tp", "TransTopic", "The topic name")
	flag.Uint64Var(&startSeq, "seq", 0, "Start at sequence no.")
	flag.Parse()

	//log.SetOutLevel(log.LvlInfo)
	log.SetOutLevel(log.LvlInfo)

	i = 0
	serviceUtil := su.GetInstanceWithOptions()

	qbh.SetDelayMsgDir("data/sub")
	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID the QBus client ID to connect with.
	// (must be deferent for every process, else will be conflict)
	qbh.Init(serviceUtil, clientID)

	// if no set, default maxNumber is: 1
	qbh.SetMaxCurrHandleMsgNumber(20)

	if startSeq == 0 {
		qbh.SubTopic(topic, "gp-1", "abc", 50, Msgcb)
	} else {
		qbh.SubTopicAt(topic, "gp-1", "abc", 1, startSeq, Msgcb)
	}

	cleanupDone = make(chan bool)
	waitSubToExit()
	<-cleanupDone
}

func waitSubToExit() {
	subC = make(chan os.Signal, 1)
	signal.Notify(subC, os.Interrupt, os.Kill)

	go func() {
		for _ = range subC {
			log.Info("get signal:")
			// just close connection, no need unsubscribe !
			qbh.CloseConn()

			cleanupDone <- true
		}
	}()
}
