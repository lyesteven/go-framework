package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	qbh "github.com/lyesteven/go-framework/common/NatsHelper"
	member "github.com/lyesteven/go-framework/common/NatsHelper/example/protobf"
	log "github.com/lyesteven/go-framework/common/new_log15"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"time"
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
	meta, _ := msg.Metadata()
	var recv_data member.AllPerson

	i++
	err := proto.Unmarshal(msg.Data, &recv_data)
	if err != nil {
		log.Error("Mashal data error:", "err", err.Error())
	}

	log.Info(fmt.Sprintf("Get Subj:[%s] message, StreamSeq=%v ConsumerSeq=%v\n", msg.Subject, meta.Sequence.Stream, meta.Sequence.Consumer))
	for no, pers := range recv_data.Per {
		log.Info(fmt.Sprintf("\tperson[%d]: ID:[%d]:  Name:%s", no, pers.Id, pers.Name))
	}
}

func waitToExit() {
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for _ = range c {
			log.Info("get signal:")
			qbh.CloseConn()

			cleanupDone <- true
		}
	}()
}

func main() {
	var startSeq uint64

	flag.Uint64Var(&startSeq, "seq", 0, "Start at sequence no.")
	flag.Parse()
	log.SetOutLevel(log.LvlInfo)
	//log.SetOutLevel(log.LvlError)
	i = 0
	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, else will be conflict)
	qbh.Init("nats://127.0.0.1:4222")

	// if no set, default maxNumber is: 1
	qbh.SetMaxCurrHandleMsgNumber(10)

	time.Sleep(1 * time.Second)

	if startSeq == 0 {
		// SubTopic(topic, qgroup, durable string, hdnum int, msgcb stan.MsgHandler)
		// @topic   the subject to listen
		// @qgroup  Queue group name
		//        (every one in same queue will consume msgs alternatly for this subject )
		// @durable  Durable subscriber name  (for persistence)
		// @hdnum	 goroutine number
		//        (the numbers to join qgroup, means can handle numbers msg at the same time )
		// @msgcb	 callback handler func
		qbh.SubTopic("hello-pb", "hgp-1", 1, Msgcb)
	} else {
		// SubTopicAt(topic, qgroup, durable string, hdnum int, startSeq uint64, msgcb stan.MsgHandler)
		// @startSeq  start position where start to get the messages
		//		 (NOTICE:  this parameter can act only at THE FIRST time in one qgroup or durable !!!)
		//
		// others parameters same as SubTopic()
		qbh.SubTopicAt("hello-pb", "hgp-1", 1, startSeq, Msgcb)
	}

	cleanupDone = make(chan bool)
	waitToExit()
	<-cleanupDone
}
