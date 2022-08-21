package main

import (
	"flag"
	"syscall"

	//"fmt"
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"github.com/hashicorp/consul/api"
	"github.com/nats-io/go-nats-streaming"
	log "github.com/xiaomi-tc/log15"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	i           int
	c           chan os.Signal
	cleanupDone chan bool

	lastSeq		uint64


	clientID, topic string
	kv *api.KV
)


var sUtil *su.ServiceUtil

var Msgcb stan.MsgHandler = func(msg *stan.Msg) {
	//i++
	if lastSeq == msg.Sequence {
		// 重复消息
		log.Info("重复消息[" + strconv.Itoa(i) + "]: msg="+string(msg.Data))
		return
	}
	lastSeq = msg.Sequence
	log.Info("callback[" + strconv.Itoa(i) + "]: msg="+string(msg.Data) + "  --seq:" + strconv.FormatUint(msg.Sequence,10))

	pair, meta, err := kv.Get(clientID, nil)
	if err != nil {
		// consul should not be error !
		log.Error("consul get kv error! in publishDelayMsg(), will retry 1 second later")
		time.Sleep(1 * time.Second)
	}

	if pair != nil {
		i,_ = strconv.Atoi(string(pair.Value))
		i++
		log.Info("CAS:","i",i)
		p := &api.KVPair{Key: clientID, Value: []byte(strconv.Itoa(i))}
		// change kv in consul. if failed will do nothing, and have another chance one second after
		p.ModifyIndex = meta.LastIndex
		kv.CAS(p, nil)
	} else {
		i = 1  // other process can reclean count number by clear KV
		log.Info("Put:","i",i)
		p := &api.KVPair{Key: clientID, Value: []byte(strconv.Itoa(i))}
		// first time, put new kv in consul
		kv.Put(p, nil)
	}

	//pair,_,_ = kv.Get(clientID, nil)
	//log.Info(clientID+" kv value:" + string(pair.Value))
}

func waitToExit() {
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP,syscall.SIGUSR1)
	go func() {
		for v := range c {
			log.Info("get signal:" + v.String())
			// just close connection, no need unsubscribe !!
			qbh.CloseConn()

			cleanupDone <- true
		}
	}()
}

func main() {

	flag.StringVar(&clientID, "id", "MyTestSubID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic, "tp", "hello", "The topic name")
	flag.Parse()

	// Open log file
	h, _ := log.FileHandler("./subTest.log" ,log.LogfmtFormat())
	log.Root().SetHandler(h)

	//log.SetOutLevel(log.LvlInfo)
	log.SetOutLevel(log.LvlDebug)

	i = 0

	log.Info("subTestTool start!---- topic:" + topic + ", clientID:"+clientID)
	sUtil := su.GetInstanceWithOptions(su.WithConsulSrvAddr("127.0.0.1:18500"))
	qbh.SetDelayMsgDir("./data/sub_delayMsg")
	qbh.Init(sUtil, clientID)
	kv = qbh.ConsulClient.KV()

	qbh.SubTopic(topic, "gp-1", "abc", 1, Msgcb)

	cleanupDone = make(chan bool)
	waitToExit()
	<-cleanupDone
}

