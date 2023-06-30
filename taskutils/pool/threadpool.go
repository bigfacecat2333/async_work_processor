package pool

import (
	"fmt"
	"sync"
)

type Job struct {
	ID  int
	Msg string
}

type ThreadPool struct {
	NumWorkers   int
	NumMaxQueues int
	jobs         chan Job
	results      chan int
	wg           sync.WaitGroup
}

func NewThreadPool(numWorkers, numMaxQueues int) *ThreadPool {
	return &ThreadPool{
		NumWorkers:   numWorkers,
		NumMaxQueues: numMaxQueues,
		jobs:         make(chan Job, numMaxQueues),
		results:      make(chan int),
	}
}

func (tp *ThreadPool) Start() {
	for i := 1; i <= tp.NumWorkers; i++ {
		tp.wg.Add(1)
		go tp.worker(i)
	}
}

func (tp *ThreadPool) worker(id int) {
	for job := range tp.jobs {
		fmt.Printf("Worker %d processing job %d\n", id, job.ID)
		// 模拟处理任务
		// ...
		tp.results <- job.ID // 将任务ID发送到结果通道
	}
	tp.wg.Done()
}

func (tp *ThreadPool) AddJob(job Job) {
	tp.jobs <- job
}

func (tp *ThreadPool) Wait() {
	close(tp.jobs)
	tp.wg.Wait()
	close(tp.results)
}

func (tp *ThreadPool) GetResults() <-chan int {
	return tp.results
}

func main() {
	const (
		NumWorkers   = 3   // 线程池中的工作协程数量
		NumJobs      = 10  // 要处理的任务数量
		NumMaxQueues = 100 // 任务队列的最大容量
	)

	tp := NewThreadPool(NumWorkers, NumMaxQueues)
	tp.Start()

	for i := 1; i <= NumJobs; i++ {
		tp.AddJob(Job{ID: i, Msg: fmt.Sprintf("Job %d", i)})
	}

	tp.Wait()

	for result := range tp.GetResults() {
		fmt.Printf("Job %d finished\n", result)
	}
}
