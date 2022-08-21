package main

import (
	"fmt"
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"github.com/nats-io/go-nats-streaming"
	log "github.com/xiaomi-tc/log15"
	"os"
	"os/signal"
	"flag"
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
var Msgcb stan.MsgHandler = func(msg *stan.Msg) {
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
	var clientID, topic string

	flag.StringVar(&clientID, "id", "MyTestSubID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic, "tp", "hello", "The topic name")
	flag.Uint64Var(&startSeq, "seq", 0, "Start at sequence no.")
	flag.Parse()

	//log.SetOutLevel(log.LvlInfo)
    log.SetOutLevel(log.LvlInfo)

	i = 0
	serviceUtil := su.GetInstanceWithOptions()
	//serviceUtil := su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))

	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, else will be conflict)
	qbh.Init(serviceUtil, clientID)

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

	qbh.SubTopicWithManualAck(topic, "gp-1", "abc", 1, Msgcb)

	cleanupDone = make(chan bool)
	waitToExit()
	<-cleanupDone

}
