package main

import (
	member "./protobf"
	su "gworld/git/GoFrameWork/common/service_util"
	qbh "gworld/git/GoFrameWork/common/utils/QBusHelper"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

func main() {
	serviceUtil := su.GetInstanceWithOptions()
	// Init(sutil *su.ServiceUtil, clientID string)
	// @clientID    the QBus client ID to connect with.
	//	  (must be deferent for every process, else will be conflict)
	qbh.Init(serviceUtil, "MyTestPubPbID")

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
		log.Fatalf("Marshal data error:", err)
	}

	// publish msg
	// Publish(topic string, pubData []byte)
	// @topic    the subject to publish
	// @pubData
	for i:=0; i<2; i++ {
		err = qbh.Publish("hello-pb", send_data)
		if err != nil {
			log.Printf("Error during publish: %v\n", err)
		}
		time.Sleep(1*time.Second)
	}

	qbh.CloseConn()
}
