package scheduler

import (
	log "github.com/xiaomi-tc/log15"
	"golang.org/x/net/context"
	"os"
	"strconv"
)

const (
	defaultWorkerNum    = 200
	defaultJobChanSize  = 10000
	defaultWorkerPrefix = "worker-"
)

var (
	maxWorkerNum   int
	maxJobChanSize int

	jobQueue chan Job
)

func init() {
	strWorkerNum := os.Getenv("maxWorkerNum")
	strJobChanSize := os.Getenv("maxJobChanSize")

	maxWorkerNum, _ = strconv.Atoi(strWorkerNum)
	if maxWorkerNum <= 0 {
		maxWorkerNum = defaultWorkerNum
	}

	maxJobChanSize, _ = strconv.Atoi(strJobChanSize)
	if maxJobChanSize <= 0 {
		maxJobChanSize = defaultJobChanSize
	}

	jobQueue = make(chan Job, maxJobChanSize)
}

type Dispatcher struct {
	ctx    context.Context
	cancel context.CancelFunc

	workerNum  int
	workerPool chan chan Job
	quit       chan struct{}
}

func NewDispatcher(workerNum int) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	pool := make(chan chan Job, workerNum)
	return &Dispatcher{
		ctx:        ctx,
		cancel:     cancel,
		workerPool: pool,
		workerNum:  workerNum,
		quit:       make(chan struct{}),
	}
}

func (d *Dispatcher) Start() {
	for i := 0; i < d.workerNum; i++ {
		workerID := defaultWorkerPrefix + strconv.Itoa(i+1)
		worker := NewWorker(d.ctx, workerID, d.workerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-jobQueue:
			jobChan := <-d.workerPool
			jobChan <- job
		case <-d.quit:
			log.Debug("Dispatcher quit")
			return
		}
	}
}

// Notice: this is blocking calls
func (d *Dispatcher) Add(job Job) {
	jobQueue <- job
}

func (d *Dispatcher) Stop() {
	go func() {
		d.quit <- struct{}{}
	}()

	// notify worker quit
	d.cancel()
}
