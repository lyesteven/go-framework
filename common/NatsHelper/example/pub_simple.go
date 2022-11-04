package main

import (
	"bufio"
	"flag"
	qbh "github.com/lyesteven/go-framework/common/NatsHelper"
	log "github.com/lyesteven/go-framework/common/new_log15"
	"os"
	"os/signal"
	"strconv"
)

var c chan os.Signal

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
	var topic, tp string
	flag.StringVar(&topic, "tp", "hello", "The topic name")
	flag.Parse()

	log.SetOutLevel(log.LvlInfo)
	//log.SetOutLevel(log.LvlDebug)

	qbh.SetDelayMsgDir("data/pub")

	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, otherwise will be conflict)
	qbh.Init("nats://127.0.0.1:4222")

	waitToExit()

	j := 0
	topic2 := "nihao"
	for {
		log.Info("Press any key to send:")
		readstring := bufio.NewReader(os.Stdin)
		readstring.ReadString('\n')

		for i := 0; i < 2; i++ {
			if i == 0 {
				tp = topic
			} else {
				tp = topic2
			}
			msg := "msg-" + tp + "-" + strconv.Itoa(j)
			j++
			data := []byte(msg)
			// publish msg
			// Publish(topic string, pubData []byte)
			// @topic    the subject to publish
			// @pubData

			if err := qbh.Publish(tp, data); err != nil {
				log.Error(" \""+string(data[:])+"\"", "error=", err)
			} else {
				log.Info(msg)
			}
		}
	}
}
