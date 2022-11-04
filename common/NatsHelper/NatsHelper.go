package NatsHelper

import (
	"errors"
	"os"
	"path"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/xiaomi-tc/boltdb"
	log "github.com/xiaomi-tc/log15"
)

// type subTopic struct {
// 	topic     string          // topic to subscribt
// 	qgroup    string          // queue group. If no optional durable name nor binding options are specified, the queue name will be used as a durable name
// 	hdnum     int             // the goroutime number for handler
// 	manualAck bool            // is need respose ack in handler. default should be false
// 	msgcb     nats.MsgHandler // call back function. will be called when received a msg from the topic channel
// }

type reqCmd struct {
	cmd       string
	topic     string // topic to subscribt
	qgroup    string // queue group. If no optional durable name nor binding options are specified, the queue name will be used as a durable name
	hdnum     int    // the goroutime number for handler
	startseq  uint64 // offset where to start to get the message from MQ.  0 means from new
	manualAck bool   // true means handler response ACK to MQ. default is false
	pubData   []byte // publish data
	msgcb     nats.MsgHandler
}

type delaymsg struct {
	Topic string `json:"topic,omitempty"`
	Msg   []byte `json:"msg,omitempty"`
}

const (
	topicPrefix             = "natsHelper" // jetstream 订阅"natsHelper.>"主题, 所有publish 和 sub的主体都会有 'natsHelper.'前缀
	delaymsgDir, delaymsgDB = "data", "dlyMsg.db"
	MaxAsyncPubOnce         = 1000
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary // Unmarshal
var (
	NatsConn    *nats.Conn
	JSContext   nats.JetStreamContext
	natsAddr    string
	isConnected = false //是否已经连接 nats-server
	//tryCount    = 1
	subTopicMap = make(map[string]reqCmd) // 保存所有sub 订阅的主题. (用于重连)

	dlyMsgDir   string = delaymsgDir
	maxInFlight int    = 1
	bInit              = false
	chRequest          = make(chan reqCmd)
	chResponse         = make(chan struct {
		string
		error
	})

	dlyMsgDB    *boltdb.DB
	dlyMsgLists *boltdb.List
	//diskQ gdq.Interface
	//currQ gdq.Interface

	inMsgHdlCount, outMsgHdlCount uint32 = 0, 0
)

func openDlyMsgLists() error {

	if dlyMsgDir == "" {
		dlyMsgDir = delaymsgDir
	}
	dbName := dlyMsgDir + "/" + delaymsgDB

	if _, err := os.Stat(dlyMsgDir); err != nil {
		log.Debug("path " + dlyMsgDir + " not exists!  Will create it ")
		err = os.MkdirAll(dlyMsgDir, os.ModePerm)

		if err != nil {
			log.Error("Error creating directory: " + err.Error())
			return err
		}
	}

	var err error
	dlyMsgDB, err = boltdb.Open(dbName, 0644, nil)
	if err != nil {
		log.Error("Open delay message DB:" + dbName + " , error:" + err.Error())
		return err
	}

	bucket, _ := dlyMsgDB.Bucket([]byte("delayMsgs"))
	log.Debug("create dlyMsg list(msgs)")
	dlyMsgLists, err = bucket.List([]byte("msgs"))
	if err != nil {
		log.Error("bucket.List(msgs)error:" + err.Error())
		return err
	}
	return nil
}

func SetMaxCurrHandleMsgNumber(maxNumber int) {
	if maxNumber > 0 {
		if maxNumber <= 5 {
			maxInFlight = maxNumber
		} else {
			maxInFlight = 5
		}
		return
	}
	maxInFlight = 0
}

// only for debug
func SetDelayMsgDir(dir string) {
	inputDir := path.Clean(dir)
	if len(inputDir) > 0 && inputDir[0] != '.' {
		dlyMsgDir = inputDir
	}
}

// jetstream 订阅"natsHelper.>"主题, 所有publish 和 sub的主体都会有 'natsHelper.'前缀
func Init(url string) error {
	// run Init() only once
	if !bInit {
		// 设置日志级别
		log.SetOutLevel(log.LvlInfo)
		//log.SetOutLevel(log.LvlDebug)

		var err error
		natsAddr = url // 保存nats-server 的地址，用于自动重连
		// 连接Nats服务器
		NatsConn, err = nats.Connect(natsAddr)
		if err != nil {
			log.Info("nats connect " + natsAddr + " error: " + err.Error())
			return err
		}

		JSContext, err = NatsConn.JetStream()
		if err != nil {
			log.Info("nats JetStream() error: " + err.Error())
			return err
		}

		_, err = JSContext.AddStream(&nats.StreamConfig{
			Name:     "NatsHelper",
			Subjects: []string{topicPrefix + ".>"},
		})
		if err != nil {
			log.Info("AddStream error: " + err.Error())
		}

		isConnected = true

		SetMaxCurrHandleMsgNumber(1)

		// boltdb, delay message list
		err = openDlyMsgLists()
		if err != nil {
			os.Exit(-1)
		}

		log.Info("")
		log.Info("Init()----------------------------------newest qbh version--------------")

		go handleDelayPubMsg()
		go manageQBusConnection()

		bInit = true
	}

	return nil
}

func manageQBusConnection() {
	log.Info("============= go manageQBusConnection")

	var req reqCmd
	//t := time.NewTicker(time.Second * 5)
	for {
		select {
		// case <-t.C:
		// check QBus chang every 5 second
		case req = <-chRequest:
			switch req.cmd {
			case "subTopic":
				// subscribe one topic
				//err := subscribeTopic(req.dcID,req.topic, req.qgroup, req.durable, req.hdnum, req.startseq, req.msgcb)
				err := subscribeTopic(req)
				chResponse <- struct {
					string
					error
				}{"", err}
			case "pubMsg":
				err := publish(req.topic, req.pubData)
				chResponse <- struct {
					string
					error
				}{"", err}
			}
		}
	}
}

func SubTopic(topic, qgroup string, hdnum int, msgcb nats.MsgHandler) error {
	// default start from seq:1. (Will be effective at the first time, and will be the same as 0 frome the second time)
	return subTopicAt(topic, qgroup, hdnum, 0, false, msgcb)
}

func SubTopicAt(topic, qgroup string, hdnum int, startSeq uint64, msgcb nats.MsgHandler) error {
	return subTopicAt(topic, qgroup, hdnum, startSeq, false, msgcb)
}

func SubTopicWithManualAck(topic, qgroup string, hdnum int, msgcb nats.MsgHandler) error {
	return subTopicAt(topic, qgroup, hdnum, 0, true, msgcb)
}

func subTopicAt(topic, qgroup string, hdnum int, startSeq uint64, isManualAck bool, msgcb nats.MsgHandler) error {
	req := reqCmd{cmd: "subTopic", topic: topic,
		qgroup: qgroup, hdnum: hdnum, startseq: startSeq,
		manualAck: isManualAck, msgcb: msgcb}
	chRequest <- req
	ret := <-chResponse
	return ret.error
}

//func subscribeTopic(dcID ComMessage.CBSType,topic, qgroup, durable string, hdnum int, startSeq uint64, msgcb stan.MsgHandler) error {
func subscribeTopic(req reqCmd) error {
	var (
		//qsc *nats.Subscription
		err error
	)
	// 保存订阅的相关信息 (用于断线后重连处理)
	if _, ok := subTopicMap[req.topic]; !ok {
		subTopicMap[req.topic] = req
	}

	if !isConnected {
		log.Error("QueueSubscribe Error: does not connected !")
		return errors.New("does not connected")
	}
	// get input data
	topic := topicPrefix + "." + req.topic
	qgroup := req.qgroup
	hdnum := req.hdnum
	startSeq := req.startseq
	isManualAck := req.manualAck
	msgcb := req.msgcb

	// 3. sub topic on every QBus node with proper order
	var Msgcb nats.MsgHandler = func(msg *nats.Msg) {
		atomic.AddUint32(&inMsgHdlCount, 1)
		msgcb(msg)
		atomic.AddUint32(&outMsgHdlCount, 1)
	}

	startOpt := nats.DeliverAll()
	//startOpt := stan.StartAt(pb.StartPosition_NewOnly)
	if startSeq != 0 {
		startOpt = nats.StartSequence(startSeq)
	}

	// get node_list from consul
	for i := 0; i < hdnum; i++ {
		if maxInFlight > 0 {
			if conInfo, err := JSContext.ConsumerInfo("NatsHelper", qgroup); err == nil && conInfo.Config.MaxAckPending != maxInFlight {
				conInfo.Config.MaxAckPending = maxInFlight
				JSContext.UpdateConsumer("NatsHelper", &conInfo.Config)
			}
			if isManualAck {
				_, err = JSContext.QueueSubscribe(topic, qgroup, Msgcb, startOpt, nats.MaxAckPending(maxInFlight), nats.ManualAck())
			} else {
				_, err = JSContext.QueueSubscribe(topic, qgroup, Msgcb, startOpt, nats.MaxAckPending(maxInFlight))
				//_, err = JSContext.QueueSubscribe(topic, qgroup, Msgcb)
			}
		} else {
			// unlimit handle msgs one time
			if isManualAck {
				_, err = JSContext.QueueSubscribe(topic, qgroup, Msgcb, startOpt, nats.ManualAck())
			} else {
				_, err = JSContext.QueueSubscribe(topic, qgroup, Msgcb, startOpt)
			}
		}
		//qsc, err := ncAddr.nsConn.QueueSubscribe(topic, qgroup, msgcb, startOpt)
		if err != nil {
			log.Error("QueueSubscribe Error in subscribeTopic()", "error", err)
			continue
		}
	}
	return nil
}

func Publish(topic string, pubData []byte) error {
	req := reqCmd{cmd: "pubMsg", topic: topic, pubData: pubData}
	chRequest <- req
	ret := <-chResponse
	return ret.error
}

func publish(topic string, pubData []byte) error {
	topic = topicPrefix + "." + topic
	log.Debug("publish()", "topic:", topic, "pubData", string(pubData))
	// log.Info("publish()", "topic:", topic, "pubData", string(pubData), "isConnected", isConnected)

	if isConnected {
		if _, err := JSContext.Publish(topic, pubData); err == nil {
			// success
			log.Debug("published OK!", "topic:", topic)
			return nil
		} else { // else
			isConnected = false
			NatsConn.Close()
			//log.Info("       in publish() --7 fail")
		}
	} else {
		dlyMsgLen, err := dlyMsgLists.Len()
		if err != nil {
			// if error, continue send current pub msg
			log.Error("get dlymsg length error, err:", err.Error())
			dlyMsgLen = 0
		}
		if dlyMsgLen == 0 && tryPubOneMessage(topic, pubData) {
			// pub message success
			log.Debug("First published OK!", "topic:", topic)
			isConnected = true
			return nil
		}
	}

	// delay send
	// Put msg into diskQueue, which will be handled by handleDelayPubMsg()
	log.Info("publish() -- delayPubMsg", "topic:", topic)
	delayMsg := new(delaymsg)
	delayMsg.Topic = topic
	delayMsg.Msg = pubData
	if msgValue, err := json.Marshal(delayMsg); err == nil {
		//if err := diskQ.Put(msgValue); err != nil {
		if err := dlyMsgLists.RPush([]byte(msgValue)); err != nil {
			log.Error("disk queue put Error:" + err.Error() + " Topic[" + topic + "] msg[" + string(pubData) + "]")
		} else {
			len, _ := dlyMsgLists.Len()
			log.Info("publish() delyMsgList", "length:", len)
		}
		//log.Debug("Put Into Msg:topic["+topic +"] msg[" +string(pubData)+"]")
	}

	return errors.New("publish will delay")
}

func handleDelayPubMsg() {
	var (
		dlyMsg delaymsg
	)
	log.Info("============= go handleDelayPubMsg")

	// watch diskQueue, and publish msgs until nil if have isConnected
	for {
		// dlyMsg处理
		if checkAndReconnect() {
			ret := false
			dlylist := dlyMsgLists
			length, err := dlylist.Len()
			// log.Info("handleDelayPubMsg() --1", "dlylist.Len()", length)
			if err == nil && length > 0 {
				// get the head msg
				queueOut, _ := dlylist.Index(0)

				if err = json.Unmarshal(queueOut, &dlyMsg); err == nil {
					ret = tryPubOneMessage(dlyMsg.Topic, []byte(dlyMsg.Msg))
				} else {
					// data maybe error, skip it
					log.Error("Unmarshal dlyMsg error: " + err.Error())
				}

				if ret {
					// publish OK, no need keep the send msg
					// log.Info("handleDelayPubMsg()  publish ok, next one ---" + ComMessage.CBSType(i).String())
					dlylist.LPop()

					// Async Publish other messages  ============ BEGIN ====================
					var isContinue = true
					len, err := dlylist.Len()

					//log.Info("handleDelayPubMsg(),  --2 publish all dly msg", "len", len)
					for ; err == nil && len > 0; len, err = dlylist.Len() {
						count := len
						if len > MaxAsyncPubOnce {
							count = MaxAsyncPubOnce
						}

						err = dlylist.Range(0, count-1, func(i int64, value []byte, quit *bool) {
							if err = json.Unmarshal(value, &dlyMsg); err == nil {
								// log.Info("handleDelayPubMsg() --3", "topic", dlyMsg.Topic, "msg", dlyMsg.Msg)
								// Publish messages asynchronously.
								_, err = JSContext.PublishAsync(dlyMsg.Topic, []byte(dlyMsg.Msg))
								if err != nil {
									log.Error("Error during async publish: " + err.Error())
									isContinue = false
								}
							} else {
								// data maybe error, skip it
								log.Error("Error during json.Unmarshal: " + err.Error())
							}
						})

						select {
						case <-JSContext.PublishAsyncComplete():
						case <-time.After(5 * time.Second):
							log.Error("PublishAsync timeout")
							isContinue = false
						}

						if !isContinue {
							break
						}
						dlylist.LBatchDelete(count)
					}
					//log.Debug("Total Async Publish:" + strconv.Itoa(total))
					// =============================== END ========================================
				} else {
					// failed, stop publish
					//log.Info("handleDelayPubMsg()  failed,stop pub ---" + ComMessage.CBSType(i).String())
				}
			}
		}

		//log.Debug("queue is Empty, Sleep 1 second")
		time.Sleep(2 * time.Second)
	}
}

func tryPubOneMessage(topic string, pubData []byte) bool {
	log.Debug("in tryPubOneMessage()", "topic:", topic)

	// msgHeader := make(nats.Header)
	// msgId := fmt.Sprintf("%d", tryCount)
	// msgHeader.Add("Nats-Msg-Id", msgId)
	// msg := &nats.Msg{Subject: topic, Header: msgHeader, Data: []byte(pubData)}
	// if _, err := JSContext.PublishMsg(msg); err == nil {
	if checkAndReconnect() {
		if _, err := JSContext.Publish(topic, pubData); err == nil {
			log.Debug("in tryPubOneMessage() success", "topic:", topic)
			log.Info("in tryPubOneMessage() success", "topic:", topic)
			//tryCount++
			return true
		} else {
			log.Error("tryPubOneMessage() error", "topic", topic, "error", err)
		}
	}

	return false
}

func CloseConn() {
	log.Debug("enter CloseConn()")
	NatsConn.Close()

	// delay message lists
	dlylist := dlyMsgLists
	l1, err := dlylist.Len()
	for ; err == nil && l1 > 0; l1, err = dlylist.Len() {
		time.Sleep(50 * time.Millisecond)
		l2, _ := dlylist.Len()
		if l2 == l1 || l2 == int64(0) {
			break
		}
	}
	// l2<l1 , still publishing delay message, wait
	dlyMsgDB.Close()
}

func checkAndReconnect() bool {
	// 检查连接是否需要重连
	var err error
	if NatsConn == nil || NatsConn.IsClosed() {
		NatsConn, err = nats.Connect(natsAddr)
		if err != nil {
			//log.Info("in checkAndReconnect(), nats connect " + natsAddr + " error: " + err.Error())
			NatsConn.Close()
			return false
		}
		JSContext, err = NatsConn.JetStream()
		if err != nil {
			//log.Info("in checkAndReconnect(), nats JetStream() error: " + err.Error())
			return false
		}
		_, err = JSContext.AddStream(&nats.StreamConfig{
			Name:     "NatsHelper",
			Subjects: []string{topicPrefix + ".>"},
		})
		if err != nil {
			//log.Info("in checkAndReconnect(),AddStream error: " + err.Error())
			return false
		}
		// 重连成功后，重新订阅
		for _, req := range subTopicMap {
			subscribeTopic(req)
		}
		log.Info("in checkAndReconnect(), nats connect " + natsAddr + " success!")
	}

	return true
}
