package workqueue

import "sync"

var requestPool sync.Pool
var responsePool sync.Pool

func init() {

	requestPool = sync.Pool{
		New: func() interface{} {
			//fmt.Printf("New Request\n")
			return &WorkRequest{}
		},
	}
	responsePool = sync.Pool{
		New: func() interface{} {
			//fmt.Printf("New Response\n")
			return &WorkResponse{}
		},
	}
}

type WorkRequest struct {
	Id   int64
	Data interface{}
}

func (self *WorkRequest) Reset() {
	self.Id = 0
	self.Data = nil
}

type WorkResponse struct {
	Id   int64
	Data interface{}
	Err  error
}

func (self *WorkResponse) Reset() {
	self.Id = 0
	self.Data = nil
	self.Err = nil
}

type Worker interface {
	Start(ctx Context)
	Stop()
}

type Context interface {
	Add(chan *WorkRequest)
	Done(int64, interface{}, error)
}

type WorkerQueue chan chan WorkRequest

type Queue interface {
}
