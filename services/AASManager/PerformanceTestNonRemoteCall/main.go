package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pbs "github.com/lyesteven/go-framework/pb/AASManager"
	pb "github.com/lyesteven/go-framework/pb/AASMessage"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var cor_num, loop_num int

var waitgroup sync.WaitGroup

func runServiceCall(loop int) {

	for i := 0; i < loop; i++ {
		call()
	}
	waitgroup.Done()
}

func call() {
	conn, err := grpc.Dial("192.168.199.223:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbs.NewAASManagerClient(conn)

	req := new(pb.InvokeServiceRequest)
	req.Uid = -1

	req.TimeStamp = time.Now().Unix()
	_, err = c.InvokeServiceByAAS(context.Background(), req)
	if err != nil {
		log.Println("InvokeServiceByAAS error: %v", err)
	}
}

func main() {
	start_time := time.Now()
	cor_num, _ = strconv.Atoi(os.Args[1])
	loop_num, _ = strconv.Atoi(os.Args[2])
	log.Println(os.Args)
	for i := 1; i < cor_num; i++ {
		go runServiceCall(loop_num)
		waitgroup.Add(1)
	}

	waitgroup.Wait()

	dur_time := time.Now().Sub(start_time)
	log.Printf("goroutine num:%d  loop times:%d elasped time:  %f seconds \n", cor_num, loop_num, dur_time.Seconds())
}
