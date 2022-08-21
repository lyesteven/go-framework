package main

import (
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	log "github.com/xiaomi-tc/log15"
	"strconv"
	"flag"
	"bufio"
	"os"
	"os/signal"
)

var	c chan os.Signal

func waitToExit() {
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for _ = range c {
			// just close connection, no need unsubscribe !!
			qbh.CloseConn()
			os.Exit(1)
		}
	}()
}

func main() {
	var clientID, topic string
	var looptimes int
	flag.StringVar(&clientID, "id", "MyTestPubID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic,"tp","hello","The topic name")
	flag.IntVar(&looptimes,"loop",2,"the loop times")
	flag.Parse()

	log.SetOutLevel(log.LvlInfo)
	//log.SetOutLevel(log.LvlDebug)

	qbh.SetDelayMsgDir("data/pub")
	serviceUtil := su.GetInstanceWithOptions()
	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, otherwise will be conflict)
	qbh.Init(serviceUtil, clientID)

	waitToExit()

	j := 0
	for {
		log.Info("Press any key to send:")
		readstring := bufio.NewReader(os.Stdin)
		readstring.ReadString('\n')

		for i:= 0; i < looptimes; i++ {
			msg := "msg-" + topic + "-" + clientID + "-" + strconv.Itoa(j)
			j++
			data := []byte(msg)
			// publish msg
			// Publish(topic string, pubData []byte)
			// @topic    the subject to publish
			// @pubData
			if err := qbh.Publish(topic, data); err != nil {
				log.Error(" \""+string(data[:])+"\"", "error=", err)
			} else {
				if i%1000 ==0 {
					log.Info(msg)
				}
			}
		}
	}
}
