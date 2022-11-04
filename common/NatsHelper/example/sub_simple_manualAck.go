package main

import (
	"flag"
	"fmt"
	qbh "github.com/lyesteven/go-framework/common/NatsHelper"
	log "github.com/lyesteven/go-framework/common/new_log15"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	//"strconv"
	//"time"
)

var (
	i           int
	c           chan os.Signal
	cleanupDone chan bool
)

/*
* callback function that processes messages delivered to
*
* @ *stan.Msg define as:
*
* type Msg struct {
*	pb.MsgProto
*	Sub         Subscription
* }
* type MsgProto struct {
*	Sequence    uint64
*	Subject     string
*	Reply       string
*	Data        []byte
*	Timestamp   int64
*	Redelivered bool
*	CRC32       uint32
* }
 */
var Msgcb nats.MsgHandler = func(msg *nats.Msg) {
	i++

	//time.Sleep(time.Second * 15)
	fmt.Printf("callback[%d]: msg=%s\n", i, msg.Data)
	msg.Ack()
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

	qbh.Init("nats://127.0.0.1:4222")

	// if no set, default maxNumber is: 1
	qbh.SetMaxCurrHandleMsgNumber(5)

	// SubTopicWithManualAck(topic, qgroup, durable string, hdnum int, msgcb stan.MsgHandler)
	// @topic   the subject to listen
	// @qgroup  Queue group name
	//        (every one in same queue will consume msgs alternatly for this subject )
	// @durable  Durable subscriber name  (for persistence)
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
