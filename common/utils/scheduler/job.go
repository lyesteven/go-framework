package scheduler

type Job interface {
	Run() error
}
