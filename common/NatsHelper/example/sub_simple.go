package main

import (
	"flag"
	"fmt"
	qbh "github.com/lyesteven/go-framework/common/NatsHelper"
	log "github.com/lyesteven/go-framework/common/new_log15"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
)

var (
	i           int
	c           chan os.Signal
	cleanupDone chan bool
)

var Msgcb nats.MsgHandler = func(msg *nats.Msg) {
	i++

	//time.Sleep(time.Second * 15)
	meta, _ := msg.Metadata()
	fmt.Printf("callback[%d]: msg=%s StreamSeq=%v ConsumerSeq=%v Header=%v\n", i, msg.Data, meta.Sequence.Stream, meta.Sequence.Consumer, msg.Header)
}

func waitToExit() {
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for _ = range c {
			log.Info("get signal:")
			// just close connection, no need unsubscribe !!
			qbh.CloseConn()

			cleanupDone <- true
		}
	}()
}

func main() {
	var startSeq uint64
	var topic string

	flag.StringVar(&topic, "tp", "hello", "The topic name")
	flag.Uint64Var(&startSeq, "seq", 0, "Start at sequence no.")
	flag.Parse()

	//log.SetOutLevel(log.LvlInfo)
	log.SetOutLevel(log.LvlInfo)

	i = 0

	qbh.SetDelayMsgDir("data/sub")
	// Init(url string)
	// @url    nats server addr
	qbh.Init("nats://127.0.0.1:4222")

	// if no set, default maxNumber is: 1
	qbh.SetMaxCurrHandleMsgNumber(20)

	// SubTopic(topic, qgroup, hdnum int, msgcb stan.MsgHandler)
	// @topic   the subject to listen
	// @qgroup  Queue group name
	//        (every one in same queue will consume msgs alternatly for this subject )
	// @hdnum	 goroutine number
	//        (the numbers to join qgroup, means can handle numbers msg at the same time )
	// @msgcb	 callback handler func
	topic2 := "nihao"
	if startSeq == 0 {
		qbh.SubTopic(topic, "gp-1", 50, Msgcb)
		qbh.SubTopic(topic2, "gp-2", 50, Msgcb)
	} else {
		qbh.SubTopicAt(topic, "gp-1", 1, startSeq, Msgcb)
		qbh.SubTopicAt(topic2, "gp-2", 1, startSeq, Msgcb)
	}

	cleanupDone = make(chan bool)
	waitToExit()
	<-cleanupDone
}
