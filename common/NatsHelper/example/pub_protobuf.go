package main

import (
	"github.com/golang/protobuf/proto"
	qbh "github.com/lyesteven/go-framework/common/NatsHelper"
	member "github.com/lyesteven/go-framework/common/NatsHelper/example/protobf"
	log "github.com/lyesteven/go-framework/common/new_log15"
	"time"
)

func main() {

	log.SetOutLevel(log.LvlInfo)
	//log.SetOutLevel(log.LvlDebug)
	qbh.SetDelayMsgDir("data/sub")
	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, else will be conflict)
	qbh.Init("nats://127.0.0.1:4222")

	// create protobuf data to send
	str_msg := "name-1"
	p1 := member.Person{
		Id:   *proto.Int32(1001),
		Name: *proto.String(str_msg),
	}
	str_msg = "name-2"
	p2 := member.Person{
		Id:   1002,
		Name: str_msg,
	}
	all_p := member.AllPerson{
		Per: []*member.Person{&p1, &p2},
	}

	send_data, err := proto.Marshal(&all_p)
	if err != nil {
		log.Error("Marshal data", "error:", err.Error())
	}

	// publish msg
	// Publish(topic string, pubData []byte)
	// @topic    the subject to publish
	// @pubData
	for i := 0; i < 2; i++ {
		err = qbh.Publish("hello-pb", send_data)
		if err != nil {
			log.Error("Error during publish:", "error", err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	qbh.CloseConn()
}
