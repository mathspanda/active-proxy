package pool

import (
	"net/http"
)

type ProxyTaskResult struct {
	Resp  *http.Response
	Error error
}

type ProxyTask struct {
	Req      *http.Request
	RespChan chan ProxyTaskResult
}

type ProxyTaskPool struct {
	taskChan chan ProxyTask // accept task
	doChan   chan int       // limit task numbers

	LimitTaskNum int
}

func NewProxyTaskPool(maxTaskNum int) (*ProxyTaskPool, error) {
	pool := &ProxyTaskPool{LimitTaskNum: maxTaskNum}
	pool.taskChan = make(chan ProxyTask, maxTaskNum)
	pool.doChan = make(chan int, maxTaskNum)
	return pool, nil
}

func (pool *ProxyTaskPool) Push(request *http.Request) <-chan ProxyTaskResult {
	task := ProxyTask{Req: request, RespChan: make(chan ProxyTaskResult)}
	pool.doChan <- 1
	pool.taskChan <- task
	return task.RespChan
}

func (pool *ProxyTaskPool) Do() {
	for {
		task := <-pool.taskChan
		go func(task ProxyTask) {
			httpClient := http.Client{}
			resp, err := httpClient.Do(task.Req)
			task.RespChan <- ProxyTaskResult{Resp: resp, Error: err}
			<-pool.doChan
		}(task)
	}
}
