package util

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxyTask struct {
	RespChan       chan bool
	target         string
	request        *http.Request
	responseWriter http.ResponseWriter
}

type ProxyTaskPoolInterface interface {
	Push(string, http.ResponseWriter, *http.Request) <-chan bool
	Do()
}

type ProxyTaskPool struct {
	taskChan chan ProxyTask // accept task
	doChan   chan int       // limit task numbers

	LimitTaskNum int
}

func NewProxyTaskPool(maxTaskNum int) (ProxyTaskPoolInterface, error) {
	pool := &ProxyTaskPool{LimitTaskNum: maxTaskNum}
	pool.taskChan = make(chan ProxyTask, maxTaskNum)
	pool.doChan = make(chan int, maxTaskNum)
	return pool, nil
}

func (pool *ProxyTaskPool) Push(target string, rw http.ResponseWriter, r *http.Request) <-chan bool {
	task := ProxyTask{
		RespChan:       make(chan bool),
		target:         target,
		request:        r,
		responseWriter: rw,
	}
	pool.doChan <- 1
	pool.taskChan <- task
	return task.RespChan
}

func (pool *ProxyTaskPool) Do() {
	for {
		task := <-pool.taskChan
		go func(task ProxyTask) {
			targetUrl, _ := url.Parse(task.target)
			reverseProxy := httputil.NewSingleHostReverseProxy(targetUrl)
			reverseProxy.ServeHTTP(task.responseWriter, task.request)
			<-pool.doChan
			task.RespChan <- true
		}(task)
	}
}
