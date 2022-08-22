package main

import (
	"fmt"
	"github.com/lyesteven/go-framework/common/utils/scheduler"
	"time"
)

type simpleJob struct {
	id int
}

// implement Job interface
func (s *simpleJob) Run() error {
	fmt.Printf("this is a simple job[%d]\n", s.id)
	return nil
}

func main() {
	dispatch := scheduler.NewDispatcher(50)
	dispatch.Start()
	defer dispatch.Stop()

	for i := 0; i < 10000; i++ {
		job := simpleJob{
			id: i,
		}
		dispatch.Add(&job)
	}

	dispatch.Stop()
	time.Sleep(1 * time.Second)
}
