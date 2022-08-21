package singlerun

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoWithPress(t *testing.T) {
	var g Group
	c := make(chan string)

	var result int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&result, 1)
		time.Sleep(10 * time.Millisecond)
		return <-c, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Do("key", fn)
			if err != nil {
				t.Errorf("Do error = %v", err)
			}

			if got, want := fmt.Sprintf("%v (%T)", v, v), "landry (string)"; got != want {
				t.Errorf("Do got = %v, want = %v", got, want)
			}
			wg.Done()
		}()
	}

	c <- "landry"
	wg.Wait()

	if got := atomic.LoadInt32(&result); got != 1 {
		t.Errorf("result = %d, want = 1", got)
	}
}
