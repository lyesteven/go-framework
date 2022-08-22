package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pbs "github.com/lyesteven/go-framework/pb/AASManager"
	pb "github.com/lyesteven/go-framework/pb/AASMessage"
	"log"
)

func main() {
	// Set up a connection to the server.

	conn, err := grpc.Dial("192.168.199.202:20000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbs.NewAASManagerClient(conn)

	//key := []byte("chuanfa@sina.com")
	//cur := time.Now().Unix()
	//ts :=strconv.FormatInt(cur,10)
	//result, err := secu.AesEncrypt([]byte("101`|`" + ts), key)
	if err != nil {
		log.Println(err)
	}

	req := new(pb.InvokeServiceRequest)
	req.Id = "-2"
	//req.Token = string(result)
	req.ServiceName = "Hub"
	req.MethodName = "WD_HUB_GetHubList"

	//req.TimeStamp = time.Now().Unix()

	r, err := c.InvokeServiceByAAS(context.Background(), req)
	if err != nil {
		log.Fatalf("could not ResultCode: %v", err)
	}

	log.Printf("ServiceResponseData: %s", r.ServiceResponseData)
}
