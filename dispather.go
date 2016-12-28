package workqueue

import "sync"

type Dispatcher struct {
	workerQueue   chan chan *WorkRequest
	workQueue     chan *WorkRequest
	workers       []Worker
	resultQueue   chan *WorkResponse
	kill          chan struct{}
	lock          sync.RWMutex
	responseQueue map[int64]chan *WorkResponse

	id  int64
	log func(str string, args ...interface{})
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		log: func(str string, args ...interface{}) {},
	}
}

func NewDispatcherWithLogger(log func(str string, args ...interface{})) *Dispatcher {
	return &Dispatcher{
		log: log,
	}
}

func (self *Dispatcher) Request(data interface{}) {

	r := requestPool.Get().(*WorkRequest)
	r.Data = data

	self.withLock(func() {
		self.id++
		r.Id = self.id
	})

	self.workQueue <- r
}

func (self *Dispatcher) withLock(fn func()) {
	self.lock.Lock()
	fn()
	self.lock.Unlock()
}

func (self *Dispatcher) RequestAndWait(data interface{}) (interface{}, error) {

	c := make(chan *WorkResponse, 1)
	r := requestPool.Get().(*WorkRequest)
	r.Data = data
	self.withLock(func() {
		self.id++
		r.Id = self.id
		self.responseQueue[r.Id] = c
	})
	//self.log("request and wait: %d\n", r.Id)

	self.workQueue <- r
	result := <-c

	self.withLock(func() {
		delete(self.responseQueue, result.Id)
	})

	d, e := result.Data, result.Err

	r.Reset()
	requestPool.Put(r)

	result.Reset()
	responsePool.Put(result)

	//self.log("request and wait done: %d\n", r.Id)
	return d, e
}

func (self *Dispatcher) Start(nworkers int, creator func(id int) Worker) {
	if self.workerQueue != nil {
		return
	}
	self.workQueue = make(chan *WorkRequest, 100)
	self.workerQueue = make(chan chan *WorkRequest, nworkers)
	self.kill = make(chan struct{})
	self.resultQueue = make(chan *WorkResponse, 100)
	self.responseQueue = make(map[int64]chan *WorkResponse)

	ctx := &context{
		workerQueue: self.workerQueue,
		workQueue:   self.workQueue,
		resultQueue: self.resultQueue,
	}

	for i := 0; i < nworkers; i++ {
		self.log("Starting worker: %d\n", i+1)
		worker := creator(i + 1)
		self.workers = append(self.workers, worker)
		worker.Start(ctx)
	}

	dispath_result := func(r *WorkResponse) {
		self.lock.RLock()
		c, ok := self.responseQueue[r.Id]
		self.lock.RUnlock()
		if !ok {
			r.Reset()
			responsePool.Put(r)
			return
		}
		c <- r
	}

	go func() {
		for {
			select {
			case work := <-self.workQueue:
				self.log("Received work request %d\n", work.Id)
				go func() {
					worker := <-self.workerQueue

					self.log("Dispatching work request: %d\n", work.Id)
					worker <- work
				}()

			case result := <-self.resultQueue:
				self.log("Received result %d\n", result.Id)
				go dispath_result(result)

			case <-self.kill:
				return
			}
		}
	}()
}

func (self *Dispatcher) Stop() {
	if self.workerQueue == nil {
		return
	}

	self.kill <- struct{}{}

	close(self.kill)
	close(self.workerQueue)
	close(self.workQueue)
	close(self.resultQueue)

}
