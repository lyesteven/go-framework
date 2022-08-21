package main

import (
	"fmt"
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	//"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/stan.go"
	log "github.com/xiaomi-tc/log15"
	"os"
	"os/signal"
	"flag"
	"sync"

	//"strconv"
	//"time"
)

var (
	i           int
	c           chan os.Signal
	cleanupDone chan bool

	countLock sync.Mutex
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
	countLock.Lock()
	i++
	countLock.Unlock()
	//fmt.Printf("callback-1 [#%d] Received on [%s]: '%s'\n", i, msg.Subject, msg)
	//message := "msg" + strconv.Itoa(i)
	//data := []byte(message)
	//if msg.Subject != "nihao" {
	//	qbh.Publish("nihao",data)
	//}

	//time.Sleep(time.Second * 15)
	fmt.Printf("callback[%d]: msg=%s\n", i, msg.Data)
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
	var clientID, topic string

	flag.StringVar(&clientID, "id", "MyTestSubID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic, "tp", "hello", "The topic name")
	flag.Uint64Var(&startSeq, "seq", 0, "Start at sequence no.")
	flag.Parse()

	//log.SetOutLevel(log.LvlInfo)
    log.SetOutLevel(log.LvlInfo)

	i = 0
	serviceUtil := su.GetInstanceWithOptions()

	qbh.SetDelayMsgDir("data/sub")
	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, else will be conflict)
	qbh.Init(serviceUtil, clientID)

	// if no set, default maxNumber is: 1
	qbh.SetMaxCurrHandleMsgNumber(20)

	// SubTopic(topic, qgroup, durable string, hdnum int, msgcb stan.MsgHandler)
	// @topic   the subject to listen
	// @qgroup  Queue group name
	//        (every one in same queue will consume msgs alternatly for this subject )
	// @durable  Durable subscriber name  (for persistence)
	// @hdnum	 goroutine number
	//        (the numbers to join qgroup, means can handle numbers msg at the same time )
	// @msgcb	 callback handler func
	if startSeq == 0 {
		qbh.SubTopic(topic, "gp-1", "abc", 50, Msgcb)
	} else {
		qbh.SubTopicAt(topic, "gp-1", "abc", 1, startSeq, Msgcb)
	}

	//test unSubtopic
	//	time.Sleep(2 * time.Second)
	//	log.Info("will UnsubTopic:","topic:","nihao","qgroup","gp-1","durable","abc")
	//	qbh.UnsubTopic("nihao","gp-1","abc")
	//	qbh.UnsubscribeAll()

	cleanupDone = make(chan bool)
	waitToExit()
	<-cleanupDone

	//qbh.PrintListAll()
	//time.Sleep(1 * time.Second)
}
