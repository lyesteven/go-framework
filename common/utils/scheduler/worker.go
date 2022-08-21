package scheduler

import (
	log "github.com/xiaomi-tc/log15"
	"golang.org/x/net/context"
)

type Worker struct {
	ctx context.Context
	id  string

	WorkerPool chan chan Job
	JobChan    chan Job
}

func NewWorker(ctx context.Context, workerID string, pool chan chan Job) *Worker {
	return &Worker{
		ctx:        ctx,
		id:         workerID,
		WorkerPool: pool,
		JobChan:    make(chan Job),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChan
			select {
			case job := <-w.JobChan:
				//log.Debug("receive job", "worker", w.id)
				if err := job.Run(); err != nil {
					log.Error("Worker", "job.Run()", err)
				}
			case <-w.ctx.Done():
				log.Debug("Worker quit", "worker", w.id)
				return
			}
		}
	}()
}
