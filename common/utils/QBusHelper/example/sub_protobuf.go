package main

import (
	member "./protobf"
	su "github.com/lyesteven/go-framework/common/service_util"
	qbh "github.com/lyesteven/go-framework/common/utils/QBusHelper"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"log"
	"os"
	"os/signal"
	"time"
	"flag"
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
var Msgcb stan.MsgHandler = func(msg *stan.Msg) {
	log.Printf("Get msg\n")
	var recv_data member.AllPerson

	i++
	err := proto.Unmarshal(msg.Data, &recv_data)
	if err != nil {
		log.Fatalln("Mashal data error:", err)
	}

	log.Printf("Get Subj:[%s] message, seq:[%v]:\n", msg.Subject, msg.Sequence)
	for no, pers := range recv_data.Per {
		log.Printf("\tperson[%d]: ID:[%d]:  Name:%s\n", no, pers.Id, pers.Name)
	}
}

func waitToExit() {
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for _ = range c {
			log.Println("get signal:")
			qbh.CloseConn()

			cleanupDone <- true
		}
	}()
}

func main() {
	var startSeq uint64

	flag.Uint64Var(&startSeq, "seq", 0, "Start at sequence no.")
	flag.Parse()
	//log.SetOutLevel(log.LvlInfo)
	log.SetOutLevel(log.LvlError)

	i = 0
	serviceUtil := su.GetInstanceWithOptions()

	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, else will be conflict)
	qbh.Init(serviceUtil, "MyTestID")

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
		qbh.SubTopic("hello-pb", "gp-1", "abc", 1, Msgcb)
	} else {
	    // SubTopicAt(topic, qgroup, durable string, hdnum int, startSeq uint64, msgcb stan.MsgHandler)
		// @startSeq  start position where start to get the messages
		//		 (NOTICE:  this parameter can act only at THE FIRST time in one qgroup or durable !!!)
		//			
		// others parameters same as SubTopic()
		qbh.SubTopicAt("hello-pb", "gp-1", "abc", 1, startSeq, Msgcb)
	}

	cleanupDone = make(chan bool)
	waitToExit()
	<-cleanupDone
}
