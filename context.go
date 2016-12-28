package workqueue

type context struct {
	workerQueue chan chan *WorkRequest
	workQueue   chan *WorkRequest
	resultQueue chan *WorkResponse
}

func (self *context) Add(w chan *WorkRequest) {
	self.workerQueue <- w
}

func (self *context) Done(id int64, d interface{}, e error) {
	r := responsePool.Get().(*WorkResponse)
	r.Id = id
	r.Data = d
	r.Err = e
	self.resultQueue <- r
}
