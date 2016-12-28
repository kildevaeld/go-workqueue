package workqueue

func NewWorker(id int, work func(request *WorkRequest) (interface{}, error)) Worker {
	// Create, and return the worker.
	worker := worker{
		id:       id,
		Work:     make(chan *WorkRequest),
		worker:   work,
		QuitChan: make(chan bool),
	}
	return &worker
}

type worker struct {
	id       int
	worker   func(request *WorkRequest) (interface{}, error)
	Work     chan *WorkRequest
	QuitChan chan bool
}

func (w *worker) ID() int {
	return w.id
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *worker) Start(ctx Context) {
	go func() {
		for {
			// Add ourselves into the worker queue.
			ctx.Add(w.Work)
			select {
			case work := <-w.Work:
				// Receive a work request.

				result, err := w.worker(work)

				ctx.Done(work.Id, result, err)

			case <-w.QuitChan:
				// We have been asked to stop.
				//w.d.log("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

func (w *worker) Stop() {
	w.QuitChan <- true
}
