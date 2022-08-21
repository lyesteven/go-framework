package main

import (
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"gworld/git/GoFrameWork/pb/ComMessage"
	log "github.com/xiaomi-tc/log15"
	"net"
	"strconv"
	"flag"
	"bufio"
	"os"
	"os/signal"
	"strings"
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

func getHostIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops:" + err.Error())
		os.Exit(1)
	}
	for _, a := range addrs {
		//判断是否正确获取到IP
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalMulticast() && !ipnet.IP.IsLinkLocalUnicast() {
			//fmt.Println(ipnet.IP.String())
			log.Info("getHostIP", "ip", ipnet)
			if ipnet.IP.To4() != nil {
				//os.Stdout.WriteString(ipnet.IP.String() + "\n")
				return ipnet.IP.String()
			}
		}
	}

	os.Stderr.WriteString("No Networking Interface Err!")
	log.Error("getHostIP", "error", err)
	os.Exit(1)
	return ""
}

func main() {
	var clientID, topic string
	flag.StringVar(&clientID, "id", "MyTestPubID", "The NATS Streaming client ID to connect with")
	flag.StringVar(&topic,"tp","hello","The topic name")
	flag.Parse()

	log.SetOutLevel(log.LvlInfo)
	//log.SetOutLevel(log.LvlDebug)
	localIP := getHostIP()
	ss := strings.Split(getHostIP(),".")
	clientID = clientID + "-" + ss[3]


	serviceUtil := su.GetInstanceWithOptions()
	qbh.SetDelayMsgDir("./data/pub")
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

		for i:= 0; i < 2; i++ {
			msg := "msg-" + localIP + "-" + clientID + "-" + strconv.Itoa(j)
			//msg := "msg-" + topic + "-" + clientID + "-" + strconv.Itoa(j)
			j++
			data := []byte(msg)
			// publish msg
			// Publish(topic string, pubData []byte)
			// @topic    the subject to publish
			// @pubData
			if err := qbh.PublishToDC(ComMessage.CBSType_Base,topic, data); err != nil {
				log.Error(" \""+string(data[:])+"\"", "error=", err)
			} else {
				log.Info(msg)
			}
		}
	}
}
